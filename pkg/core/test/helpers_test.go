package core_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestParseUnifiedQuery_EmptyJSON(t *testing.T) {
	_, err := core.ParseUnifiedQuery(backend.DataQuery{})
	if err == nil {
		t.Fatal("expected error for empty JSON")
	}
}

func TestParseUnifiedQuery_InvalidJSON(t *testing.T) {
	_, err := core.ParseUnifiedQuery(backend.DataQuery{JSON: []byte(`{`)})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseUnifiedQuery_OK(t *testing.T) {
	b := []byte(`{"product":"cdn","metric":{"value":"x","label":"X"}}`)
	qm, err := core.ParseUnifiedQuery(backend.DataQuery{JSON: b})
	if err != nil {
		t.Fatal(err)
	}
	if qm.Product != "cdn" || qm.Metric.Value != "x" {
		t.Fatalf("unexpected model: %+v", qm)
	}
}

func TestBuildStatsRequest_NoMetric(t *testing.T) {
	b, _ := json.Marshal(core.QueryModel{Product: "cdn"})
	_, err := core.BuildStatsRequest(backend.DataQuery{JSON: b})
	if err == nil {
		t.Fatal("expected error when metric is empty")
	}
}

func TestBuildStatsRequest_BandwidthRemap(t *testing.T) {
	b, _ := json.Marshal(core.QueryModel{
		Product:     "cdn",
		Metric:      core.SelectableValue{Value: "bandwidth"},
		Granularity: core.SelectableValue{Value: "1h"},
	})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	req, err := core.BuildStatsRequest(backend.DataQuery{
		JSON:      b,
		TimeRange: backend.TimeRange{From: from, To: to},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Metrics) != 1 || req.Metrics[0] != "max_bandwidth" {
		t.Fatalf("metrics: %#v", req.Metrics)
	}
	if req.Granularity != "" {
		t.Fatalf("expected empty granularity for bandwidth, got %q", req.Granularity)
	}
}

func TestBuildStatsRequest_DefaultGranularity(t *testing.T) {
	b, _ := json.Marshal(core.QueryModel{
		Product: "cdn",
		Metric:  core.SelectableValue{Value: "total_bytes"},
	})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	req, err := core.BuildStatsRequest(backend.DataQuery{
		JSON:      b,
		TimeRange: backend.TimeRange{From: from, To: from.Add(time.Hour)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Granularity != "1h" {
		t.Fatalf("granularity: %q", req.Granularity)
	}
}

func TestBuildStatsRequest_FiltersAndGrouping(t *testing.T) {
	b, _ := json.Marshal(core.QueryModel{
		Product:     "cdn",
		Metric:      core.SelectableValue{Value: "total_bytes"},
		Granularity: core.SelectableValue{Value: "5m"},
		Grouping: []core.SelectableValue{
			{Value: "region"},
			{Value: ""},
			{Value: "vhost"},
		},
		Vhosts:    " a , b ",
		Resources: "1, x , 2",
		Clients:   "3",
		Countries: "DE, FR",
		Regions:   "EU",
	})
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	req, err := core.BuildStatsRequest(backend.DataQuery{
		JSON:      b,
		TimeRange: backend.TimeRange{From: from, To: from.Add(time.Hour)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Vhosts) != 2 || req.Vhosts[0] != "a" {
		t.Fatalf("vhosts: %#v", req.Vhosts)
	}
	if len(req.Resources) != 2 || req.Resources[0] != 1 || req.Resources[1] != 2 {
		t.Fatalf("resources: %#v", req.Resources)
	}
	if len(req.Clients) != 1 || req.Clients[0] != 3 {
		t.Fatalf("clients: %#v", req.Clients)
	}
	if len(req.GroupBy) != 2 {
		t.Fatalf("groupBy: %#v", req.GroupBy)
	}
}

func TestExtractGrouping(t *testing.T) {
	got := core.ExtractGrouping([]core.SelectableValue{{Value: "a"}, {Value: ""}, {Value: "b"}})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseInts_ParseStrings(t *testing.T) {
	if got := core.ParseInts("1, foo ,2"); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("ParseInts: %#v", got)
	}
	if got := core.ParseStrings(" a , , b "); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("ParseStrings: %#v", got)
	}
}

func TestIsAllToken_ParseSelection(t *testing.T) {
	if !core.IsAllToken(" ALL ") {
		t.Fatal("expected all token")
	}
	if core.IsAllToken("x") {
		t.Fatal("expected not all")
	}
	if got := core.ParseSelection("a,b"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("ParseSelection: %#v", got)
	}
}

func TestDnsGranularityToAPI(t *testing.T) {
	if got := core.DnsGranularityToAPI(""); got != "5m" {
		t.Fatalf("empty: %q", got)
	}
	if got := core.DnsGranularityToAPI("1h"); got != "1h" {
		t.Fatalf("1h: %q", got)
	}
	if got := core.DnsGranularityToAPI("30"); got != "30s" {
		t.Fatalf("numeric: %q", got)
	}
	if got := core.DnsGranularityToAPI("weird"); got != "5m" {
		t.Fatalf("unknown: %q", got)
	}
}
