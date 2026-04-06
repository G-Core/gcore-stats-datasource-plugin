package datasource

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func sendJSONError(sender backend.CallResourceResponseSender, status int, msg string) {
	body, err := json.Marshal(map[string]string{"error": msg})
	if err != nil {
		body = []byte(`{"error":"internal error"}`)
	}
	sender.Send(&backend.CallResourceResponse{
		Status:  status,
		Body:    body,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
	})
}

func (ds *GCDataSource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	subpath := resourceSubpath(req.Path, req.URL)
	if subpath == "" {
		sendJSONError(sender, 400, "missing path")
		return nil
	}
	base := ds.BaseAPIURL()
	var upstreamPath string
	switch subpath {
	case "iam/users/me", "users/me":
		upstreamPath = subpath
	case "cdn/resources":
		upstreamPath = "cdn/resources"
	case "dns/zones":
		upstreamPath = "dns/v2/zones"
	case "fastedge/apps":
		upstreamPath = "fastedge/v1/apps"
	default:
		sendJSONError(sender, 404, "unknown path")
		return nil
	}
	query := upstreamQueryFromResourceRequest(req.Path, req.URL)
	url := base + "/" + upstreamPath + query
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, nil)
	if err != nil {
		backend.Logger.Error("failed to create resource request", "error", err, "url", url)
		sendJSONError(sender, 500, "failed to create request")
		return nil
	}
	ds.setHeaders(httpReq)
	resp, err := ds.Client.Do(httpReq)
	if err != nil {
		backend.Logger.Error("resource request failed", "error", err, "url", url)
		sendJSONError(sender, 502, "request failed")
		return nil
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, core.DefaultMaxResponseBodyBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		backend.Logger.Error("failed to read resource response body", "error", err, "url", url)
		sendJSONError(sender, 502, "failed to read response")
		return nil
	}
	if int64(len(body)) > core.DefaultMaxResponseBodyBytes {
		sendJSONError(sender, 502, "Gcore API response too large")
		return nil
	}
	status := resp.StatusCode
	if status >= 500 {
		status = 502
		sendJSONError(sender, status, "Gcore API server error. Please try again later.")
		return nil
	}
	sender.Send(&backend.CallResourceResponse{
		Status:  status,
		Body:    body,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
	})
	return nil
}
