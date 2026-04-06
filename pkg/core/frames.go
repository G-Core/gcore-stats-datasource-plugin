package core

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func ParseRFC3339(ts string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func NewTimeValueFrame(frameName, timeFieldName, valueFieldName string) *data.Frame {
	return data.NewFrame(frameName,
		data.NewField(timeFieldName, nil, []time.Time{}),
		data.NewField(valueFieldName, nil, []float64{}),
	)
}

