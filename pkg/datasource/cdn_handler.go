package datasource

import (
	"context"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func (ds *GCDataSource) queryCDN(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	payload, err := core.BuildStatsRequest(query)
	if err != nil {
		return backend.DataResponse{Error: err}
	}

	client := &edgenetwork.Client{
		RootURL: ds.BaseAPIURL(),
		APIKey:  ds.APIKey,
		HTTP:    ds.Client,
	}
	frames, err := client.QueryCDN(ctx, payload)
	if err != nil {
		return backend.DataResponse{Error: err}
	}
	return backend.DataResponse{Frames: frames}
}
