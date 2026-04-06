package datasource

import (
	"context"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func (ds *GCDataSource) queryDNS(ctx context.Context, query backend.DataQuery, qm *core.QueryModel) backend.DataResponse {
	client := &edgenetwork.Client{
		RootURL: ds.BaseAPIURL(),
		APIKey:  ds.APIKey,
		HTTP:    ds.Client,
	}
	frames, err := client.QueryDNS(ctx, qm, query.TimeRange)
	if err != nil {
		return backend.DataResponse{Error: err}
	}
	return backend.DataResponse{Frames: frames}
}
