package edgenetwork

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (c *Client) QueryFastEdge(ctx context.Context, qm *core.QueryModel, tr backend.TimeRange) ([]*data.Frame, error) {
	reqURL := c.RootURL + "/fastedge/v1/stats/app_duration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("from", tr.From.UTC().Format(time.RFC3339))
	q.Add("to", tr.To.UTC().Format(time.RFC3339))

	step := qm.Step
	if step <= 0 {
		step = 60
	}
	q.Add("step", fmt.Sprintf("%d", step))
	if qm.AppId != 0 {
		q.Add("id", fmt.Sprintf("%d", qm.AppId))
	}
	if qm.Network != "" {
		q.Add("network", qm.Network)
	}
	req.URL.RawQuery = q.Encode()

	var out core.FastEdgeResponseStats
	_, err = core.DoJSONRequest(
		c.HTTP,
		req,
		&out,
		c.setHeaders,
		core.HandleAPIError,
		http.StatusOK,
	)
	if err != nil {
		return nil, err
	}

	return c.fastEdgeTransformToFrames(&out, qm), nil
}

func (c *Client) fastEdgeTransformToFrames(response *core.FastEdgeResponseStats, qm *core.QueryModel) []*data.Frame {
	if response == nil || len(response.Stats) == 0 {
		return nil
	}
	metric := qm.FastedgeMetric
	if metric == "" {
		metric = "avg"
	}

	frame := core.NewTimeValueFrame("", "Time", "value")
	for _, s := range response.Stats {
		t, ok := core.ParseRFC3339(s.Time)
		if !ok {
			continue
		}
		val := s.Avg
		switch metric {
		case "min":
			val = s.Min
		case "max":
			val = s.Max
		case "median":
			val = s.Median
		case "perc75":
			val = s.Perc75
		case "perc90":
			val = s.Perc90
		}
		frame.AppendRow(t, val)
	}

	name := metric
	if qm.AppName != "" {
		name = qm.AppName + " " + metric
	}
	frame.Name = name
	frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})
	return []*data.Frame{frame}
}

