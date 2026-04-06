package edgenetwork

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// ParseWaapSeries is exported so `pkg/tests` can validate permissive parsing behavior.
func ParseWaapSeries(rawMap map[string][]map[string]interface{}) map[string][]core.WaapDataPoint {
	// keep the permissive behavior from earlier implementation
	stats := make(map[string][]core.WaapDataPoint, len(rawMap))
	for metricKey, rawPoints := range rawMap {
		points := make([]core.WaapDataPoint, 0, len(rawPoints))
		for _, rp := range rawPoints {
			dt := ""
			for _, key := range []string{"date_time", "datetime", "date", "timestamp", "time"} {
				if v, ok := rp[key]; ok {
					dt = core.ToString(v)
					break
				}
			}

			val := 0.0
			for _, key := range []string{"value", "count", "total", "sum", "requests"} {
				if v, ok := rp[key]; ok {
					if n, ok2 := core.ToFloat64(v); ok2 {
						val = n
					}
					break
				}
			}

			points = append(points, core.WaapDataPoint{DateTime: dt, Value: val})
		}
		stats[metricKey] = points
	}
	return stats
}

func (c *Client) QueryWAAP(ctx context.Context, qm *core.QueryModel, tr backend.TimeRange) ([]*data.Frame, error) {
	reqURL := c.RootURL + "/waap/v1/statistics/series"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	granularity := qm.WaapGranularity
	if granularity == "" {
		granularity = "1h"
	}
	metric := qm.WaapMetric
	if metric == "" {
		metric = "total_requests"
	}

	q := req.URL.Query()
	q.Add("from", tr.From.UTC().Format(time.RFC3339))
	q.Add("to", tr.To.UTC().Format(time.RFC3339))
	q.Add("granularity", granularity)
	q.Add("metrics", metric)
	req.URL.RawQuery = q.Encode()

	raw, err := core.DoJSONRequest(
		c.HTTP,
		req,
		nil,
		c.setHeaders,
		core.HandleAPIError,
		http.StatusOK,
	)
	if err != nil {
		return nil, err
	}

	var rawMap map[string][]map[string]interface{}
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return nil, err
	}

	waapResp := &core.WaapStatsResponse{Data: ParseWaapSeries(rawMap)}
	return c.waapTransformToFrames(waapResp, qm), nil
}

func (c *Client) waapTransformToFrames(stats *core.WaapStatsResponse, qm *core.QueryModel) []*data.Frame {
	var frames []*data.Frame
	if stats == nil || len(stats.Data) == 0 {
		return frames
	}
	for metricName, points := range stats.Data {
		frame := data.NewFrame(metricName,
			data.NewField("time", nil, []time.Time{}),
			data.NewField("value", nil, []float64{}),
		)
		displayName := metricName
		if qm != nil && qm.WaapLegendFormat != "" {
			displayName = qm.WaapLegendFormat
		}
		unit := "short"
		if metricName == "total_bytes" {
			unit = "kbytes"
		}
		frame.Fields[1].SetConfig(&data.FieldConfig{DisplayName: displayName, Unit: unit})
		for _, point := range points {
			t, err := time.Parse(time.RFC3339, point.DateTime)
			if err != nil {
				continue
			}
			val := point.Value
			if metricName == "total_bytes" {
				val = val / 1024
			}
			frame.AppendRow(t, val)
		}
		frames = append(frames, frame)
	}
	return frames
}

