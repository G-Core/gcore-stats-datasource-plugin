package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type IAMUser struct {
	Name string `json:"name"`
}

func (ds *GCDataSource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	base := ds.BaseAPIURL()
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/iam/users/me", nil)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	ds.setHeaders(reqHTTP)
	resp, err := ds.Client.Do(reqHTTP)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	if err := core.HandleAPIError(resp.StatusCode, raw); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}
	var user IAMUser
	if err := json.Unmarshal(raw, &user); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	name := "Unknown"
	if user.Name != "" {
		name = user.Name
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Auth OK (IAM): %s", name),
	}, nil
}
