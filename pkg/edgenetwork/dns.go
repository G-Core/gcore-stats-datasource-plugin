package edgenetwork

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (c *Client) QueryDNS(ctx context.Context, qm *core.QueryModel, tr backend.TimeRange) ([]*data.Frame, error) {
	zones, err := c.dnsResolveZones(ctx, qm)
	if err != nil {
		return nil, err
	}
	if len(zones) == 0 {
		return c.dnsTransformToFrames(nil, qm), nil
	}

	aggregated := make(map[string]float64)
	var errs []error
	for _, zoneName := range zones {
		stats, err := c.dnsFetchZoneStats(ctx, zoneName, qm, tr)
		if err != nil {
			errs = append(errs, fmt.Errorf("dns zone %q: %w", zoneName, err))
			continue
		}
		if stats == nil {
			continue
		}
		for ts, val := range stats.Requests {
			aggregated[ts] += val
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return c.dnsTransformToFrames(&core.DNSStatsResponse{Requests: aggregated}, qm), nil
}

func (c *Client) dnsResolveZones(ctx context.Context, qm *core.QueryModel) ([]string, error) {
	if z := strings.TrimSpace(qm.Zone); z != "" && !core.IsAllToken(z) {
		return core.ParseSelection(z), nil
	}
	allZones, err := c.dnsGetAllZones(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(allZones))
	for _, z := range allZones {
		out = append(out, z.Name)
	}
	return out, nil
}

func (c *Client) dnsGetAllZones(ctx context.Context) ([]core.Zone, error) {
	baseURL := c.RootURL + "/dns"
	var allZones []core.Zone
	limit := 1000
	offset := 0
	for {
		var zonesResp core.ZonesResponse
		_, err := core.DoJSON(
			ctx,
			c.HTTP,
			http.MethodGet,
			fmt.Sprintf("%s/v2/zones?limit=%d&offset=%d", baseURL, limit, offset),
			nil,
			&zonesResp,
			c.setHeaders,
			core.HandleAPIError,
			http.StatusOK,
		)
		if err != nil {
			return nil, err
		}
		allZones = append(allZones, zonesResp.Zones...)
		if len(zonesResp.Zones) < limit || (zonesResp.TotalAmount > 0 && offset+limit >= zonesResp.TotalAmount) {
			break
		}
		offset += limit
	}
	return allZones, nil
}

func (c *Client) dnsFetchZoneStats(ctx context.Context, zoneName string, qm *core.QueryModel, tr backend.TimeRange) (*core.DNSStatsResponse, error) {
	baseURL := c.RootURL + "/dns"
	reqURL := fmt.Sprintf("%s/v2/zones/%s/statistics", baseURL, url.PathEscape(zoneName))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("from", fmt.Sprintf("%d", tr.From.Unix()))
	q.Add("to", fmt.Sprintf("%d", tr.To.Unix()))
	gran := qm.DnsGranularity.Value
	if gran == "" {
		gran = "5m"
	}
	q.Add("granularity", core.DnsGranularityToAPI(gran))
	if qm.RecordType != "" && qm.RecordType != "ALL" {
		q.Add("record_type", qm.RecordType)
	}
	req.URL.RawQuery = q.Encode()

	status, raw, err := core.DoRequest(c.HTTP, req, c.setHeaders)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err := core.HandleAPIError(status, raw); err != nil {
		return nil, err
	}

	var stats core.DNSStatsResponse
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &stats); err != nil {
			return nil, err
		}
	}
	return &stats, nil
}

func (c *Client) dnsTransformToFrames(stats *core.DNSStatsResponse, qm *core.QueryModel) []*data.Frame {
	frame := data.NewFrame("Statistics",
		data.NewField("time", nil, []time.Time{}),
		data.NewField("value", nil, []float64{}),
	)

	displayName := "Requests"
	if qm != nil && qm.DnsLegendFormat != "" {
		displayName = qm.DnsLegendFormat
	} else if qm != nil {
		if qm.Zone != "" && qm.Zone != "all" {
			displayName = "Requests (" + qm.Zone + ")"
		} else {
			displayName = "Requests (All Zones)"
		}
		if qm.RecordType != "" && qm.RecordType != "ALL" {
			displayName += ", " + qm.RecordType
		}
	}
	frame.Fields[1].SetConfig(&data.FieldConfig{DisplayName: displayName, Unit: "short"})

	if stats == nil || len(stats.Requests) == 0 {
		return []*data.Frame{frame}
	}

	keys := make([]string, 0, len(stats.Requests))
	for k := range stats.Requests {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ts, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			continue
		}
		if ts > 10000000000 {
			ts = ts / 1000
		}
		frame.AppendRow(time.Unix(ts, 0), stats.Requests[k])
	}

	return []*data.Frame{frame}
}

