package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	grafanads "github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

func main() {
	if err := grafanads.Manage("gcore-stats-datasource-plugin", newDatasourceFactory(), grafanads.ManageOpts{}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start datasource: %v\n", err)
		os.Exit(1)
	}
}

func newDatasourceFactory() grafanads.InstanceFactoryFunc {
	return func(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
			jsonData = make(map[string]interface{})
		}

		url, _ := jsonData["apiUrl"].(string)
		url = strings.TrimSpace(url)
		if url == "" {
			url = "https://api.gcore.com"
		} else {
			if err := datasource.ValidateAPIBaseURL(url); err != nil {
				return nil, err
			}
		}

		apiKey := ""
		if settings.DecryptedSecureJSONData != nil {
			apiKey = settings.DecryptedSecureJSONData["apiKey"]
		}

		return datasource.NewDataSource(url, apiKey), nil
	}
}
