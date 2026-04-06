package edgenetwork

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// QueryCDN calls the CDN statistics API and returns Grafana frames.
func (c *Client) QueryCDN(ctx context.Context, req *core.StatsRequest) ([]*data.Frame, error) {
	stats, err := c.fetchStats(ctx, req)
	if err != nil {
		return nil, err
	}
	return transformStatsToFrames(stats, req), nil
}

func (c *Client) fetchStats(ctx context.Context, payload *core.StatsRequest) ([]core.StatsResponse, error) {
	var out []core.StatsResponse
	_, err := core.DoJSON(
		ctx,
		c.HTTP,
		http.MethodPost,
		c.RootURL+"/cdn/statistics/aggregate/stats",
		payload,
		&out,
		c.setHeaders,
		core.HandleAPIError,
		http.StatusOK,
		http.StatusNoContent,
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func transformStatsToFrames(stats []core.StatsResponse, req *core.StatsRequest) []*data.Frame {
	var frames []*data.Frame

	for _, s := range stats {
		var groups []string

		if s.Resource != nil {
			groups = append(groups, fmt.Sprintf("resource: %d", *s.Resource))
		}
		if s.Client != nil {
			groups = append(groups, fmt.Sprintf("client: %d", *s.Client))
		}
		if s.Region != "" {
			groups = append(groups, fmt.Sprintf("region: %s", s.Region))
		}
		if s.Vhost != "" {
			groups = append(groups, fmt.Sprintf("vhost: %s", s.Vhost))
		}
		if s.Country != "" {
			groups = append(groups, fmt.Sprintf("country: %s", s.Country))
		}
		if s.DC != "" {
			groups = append(groups, fmt.Sprintf("dc: %s", s.DC))
		}

		suffix := ""
		if len(groups) > 0 {
			suffix = " (" + strings.Join(groups, ", ") + ")"
		}

		for metric, raw := range s.Metrics {
			var points [][2]float64
			if err := json.Unmarshal(raw, &points); err != nil || len(points) == 0 {
				var scalar float64
				if err2 := json.Unmarshal(raw, &scalar); err2 != nil {
					continue
				}
				frame := data.NewFrame(metric+suffix,
					data.NewField("time", nil, []time.Time{time.Unix(reqTime(req.From), 0)}),
					data.NewField("value", nil, []float64{scalar}),
				)
				frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})
				frames = append(frames, frame)
				continue
			}

			times := make([]time.Time, 0, len(points))
			values := make([]float64, 0, len(points))
			for _, p := range points {
				ts := int64(p[0])
				if ts > 10000000000 {
					ts = ts / 1000
				}
				times = append(times, time.Unix(ts, 0))
				values = append(values, p[1])
			}
			frame := data.NewFrame(metric+suffix,
				data.NewField("time", nil, times),
				data.NewField("value", nil, values),
			)
			frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})
			frames = append(frames, frame)
		}
	}
	return frames
}

func reqTime(rfc3339 string) int64 {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return time.Now().Unix()
	}
	return t.Unix()
}

