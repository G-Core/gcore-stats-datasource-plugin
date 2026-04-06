package edgenetwork_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
)

func TestQueryCDN_OK_TimeSeries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var payload core.StatsRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Metrics) != 1 || payload.Metrics[0] != "total_bytes" {
			t.Fatalf("metrics: %#v", payload.Metrics)
		}
		if r.Header.Get("Authorization") != "APIKey secret" {
			t.Errorf("Authorization header: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`[{"metrics":{"total_bytes":[[1704067200,42],[1704067500,43]]},"region":"eu"}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{
		RootURL: srv.URL,
		APIKey:  "secret",
		HTTP:    srv.Client(),
	}
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	req := &core.StatsRequest{
		Metrics:     []string{"total_bytes"},
		Granularity: "1h",
		From:        from.Format(time.RFC3339),
		To:          from.Add(time.Hour).Format(time.RFC3339),
		Flat:        true,
	}
	frames, err := cli.QueryCDN(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: %d", len(frames))
	}
	if frames[0].Name != "total_bytes (region: eu)" {
		t.Fatalf("frame name: %q", frames[0].Name)
	}
}

func TestQueryCDN_OK_ScalarMetric(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"metrics":{"hits":99.5}}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	statsReq := &core.StatsRequest{
		Metrics: []string{"hits"},
		From:    from.Format(time.RFC3339),
		To:      from.Add(time.Hour).Format(time.RFC3339),
		Flat:    true,
	}
	frames, err := cli.QueryCDN(context.Background(), statsReq)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 || frames[0].Name != "hits" {
		t.Fatalf("got %#v", frames)
	}
}

func TestQueryCDN_NoContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	req := &core.StatsRequest{
		Metrics: []string{"total_bytes"},
		From:    time.Now().UTC().Format(time.RFC3339),
		To:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		Flat:    true,
	}
	frames, err := cli.QueryCDN(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 0 {
		t.Fatalf("expected no frames, got %d", len(frames))
	}
}

func TestQueryCDN_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn/statistics/aggregate/stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	req := &core.StatsRequest{
		Metrics: []string{"total_bytes"},
		From:    time.Now().UTC().Format(time.RFC3339),
		To:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		Flat:    true,
	}
	_, err := cli.QueryCDN(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
}
