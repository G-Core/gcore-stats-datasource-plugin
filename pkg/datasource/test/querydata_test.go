package datasource_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func testTimeRange() backend.TimeRange {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return backend.TimeRange{From: from, To: from.Add(time.Hour)}
}

func TestGCDataSource_BaseAPIURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://api.example.com", "https://api.example.com"},
		{"https://api.example.com/cdn", "https://api.example.com"},
		{"https://api.example.com/dns", "https://api.example.com"},
		{"http://localhost:8080/fastedge", "http://localhost:8080"},
	}
	for _, tt := range tests {
		ds := datasource.NewDataSource(tt.in, "")
		if got := ds.BaseAPIURL(); got != tt.want {
			t.Errorf("BaseAPIURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGCDataSource_QueryData_ParseError(t *testing.T) {
	ds := datasource.NewDataSource("http://example.com", "")
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{
			RefID:     "A",
			JSON:      []byte(`not-json`),
			TimeRange: testTimeRange(),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected parse error")
	}
}

func TestGCDataSource_QueryData_CDN_OK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected non-empty POST body")
		}
		_, _ = w.Write([]byte(`[{"metrics":{"total_bytes":[[1704067200,100]]}}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "secret")
	ds.Client = srv.Client()

	qJSON := []byte(`{"product":"cdn","metric":{"value":"total_bytes","label":"x"},"granularity":{"value":"1h","label":"1h"}}`)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error != nil {
		t.Fatal(dr.Error)
	}
	if len(dr.Frames) == 0 {
		t.Fatal("expected frames")
	}
}

func TestGCDataSource_QueryData_CDN_Unauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "bad")
	ds.Client = srv.Client()

	qJSON := []byte(`{"product":"cdn","metric":{"value":"total_bytes","label":"x"},"granularity":{"value":"1h","label":"1h"}}`)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected error")
	}
}

func TestGCDataSource_QueryData_CDN_BuildError_NoMetric(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "")
	ds.Client = srv.Client()

	qJSON := []byte(`{"product":"cdn","metric":{"value":"","label":""}}`)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected BuildStatsRequest error")
	}
}

func TestGCDataSource_QueryData_DNS_SingleZone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/myzone/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"requests":{"1704067200":99}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "k")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "dns",
		"zone":           "myzone",
		"dnsGranularity": map[string]string{"value": "5m", "label": "5m"},
		"record_type":    "ALL",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error != nil {
		t.Fatal(dr.Error)
	}
	if len(dr.Frames) != 1 {
		t.Fatalf("frames: %d", len(dr.Frames))
	}
}

func TestGCDataSource_QueryData_DNS_AllZones(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("limit") != "1000" || q.Get("offset") != "0" {
			t.Errorf("limit/offset: %v", q)
		}
		_, _ = w.Write([]byte(`{"zones":[{"name":"zone-a","id":1}],"total_amount":1}`))
	})
	mux.HandleFunc("/dns/v2/zones/zone-a/statistics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"requests":{"1704067200":3}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "k")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "dns",
		"zone":           "all",
		"dnsGranularity": map[string]string{"value": "5m", "label": "5m"},
		"record_type":    "ALL",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error != nil {
		t.Fatal(dr.Error)
	}
	if len(dr.Frames) != 1 {
		t.Fatalf("frames: %d", len(dr.Frames))
	}
}

func TestGCDataSource_QueryData_DNS_ZonesListError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"slow down"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "k")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "dns",
		"zone":           "all",
		"dnsGranularity": map[string]string{"value": "5m", "label": "5m"},
		"record_type":    "ALL",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected error from zones list")
	}
}

func TestGCDataSource_QueryData_DNS_ZoneStatsError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/z1/statistics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "k")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "dns",
		"zone":           "z1",
		"dnsGranularity": map[string]string{"value": "5m", "label": "5m"},
		"record_type":    "ALL",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected error")
	}
}

func TestGCDataSource_QueryData_FastEdge_OK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("step") != "60" {
			t.Errorf("step=%q", q.Get("step"))
		}
		if !strings.Contains(q.Get("from"), "2026") {
			t.Errorf("from: %q", q.Get("from"))
		}
		_, _ = w.Write([]byte(`{"stats":[{"time":"2026-01-01T00:00:00Z","avg":1.5,"min":1,"max":2,"median":1.5,"perc75":2,"perc90":2}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "fastedge",
		"fastedgeMetric": "avg",
		"step":           float64(60),
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error != nil {
		t.Fatal(dr.Error)
	}
	if len(dr.Frames) != 1 {
		t.Fatalf("frames: %d", len(dr.Frames))
	}
}

func TestGCDataSource_QueryData_FastEdge_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":        "fastedge",
		"fastedgeMetric": "avg",
		"step":           float64(60),
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected error")
	}
}

func TestGCDataSource_QueryData_WAAP_OK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/waap/v1/statistics/series", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("metrics") != "total_requests" {
			t.Errorf("metrics=%q", r.URL.Query().Get("metrics"))
		}
		_, _ = w.Write([]byte(`{"total_requests":[{"date_time":"2026-01-01T00:00:00Z","value":7}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":         "waap",
		"waapMetric":      "total_requests",
		"waapGranularity": "1h",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	dr := resp.Responses["A"]
	if dr.Error != nil {
		t.Fatal(dr.Error)
	}
	if len(dr.Frames) != 1 {
		t.Fatalf("frames: %d", len(dr.Frames))
	}
}

func TestGCDataSource_QueryData_WAAP_InvalidJSONBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/waap/v1/statistics/series", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "")
	ds.Client = srv.Client()

	qJSON, _ := json.Marshal(map[string]interface{}{
		"product":         "waap",
		"waapMetric":      "total_requests",
		"waapGranularity": "1h",
	})

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error == nil {
		t.Fatal("expected json unmarshal error")
	}
}

func TestGCDataSource_QueryData_DefaultProductCDN_WithRootURLStrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL+"/cdn", "")
	ds.Client = srv.Client()

	qJSON := []byte(`{"metric":{"value":"total_bytes","label":"x"},"granularity":{"value":"1h","label":"1h"}}`)

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: qJSON, TimeRange: testTimeRange()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Responses["A"].Error != nil {
		t.Fatal(resp.Responses["A"].Error)
	}
}
