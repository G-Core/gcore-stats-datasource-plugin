package edgenetwork_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func feTR() backend.TimeRange {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return backend.TimeRange{From: from, To: from.Add(time.Hour)}
}

func TestQueryFastEdge_OK_DefaultAvg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("step") != "60" {
			t.Errorf("step=%q", q.Get("step"))
		}
		if !strings.Contains(q.Get("from"), "2026") {
			t.Errorf("from=%q", q.Get("from"))
		}
		_, _ = w.Write([]byte(`{"stats":[{"time":"2026-01-01T00:00:00Z","avg":1.1,"min":1,"max":2,"median":1.5,"perc75":1.8,"perc90":2}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{Step: 60, FastedgeMetric: ""}
	frames, err := cli.QueryFastEdge(context.Background(), qm, feTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 || frames[0].Name != "avg" {
		t.Fatalf("frames: %#v", frames)
	}
}

func TestQueryFastEdge_QueryParams_AppIdAndNetwork(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("id") != "7" {
			t.Errorf("id=%q", q.Get("id"))
		}
		if q.Get("network") != "edge-net" {
			t.Errorf("network=%q", q.Get("network"))
		}
		_, _ = w.Write([]byte(`{"stats":[{"time":"2026-01-01T00:00:00Z","avg":0,"min":2,"max":2,"median":2,"perc75":2,"perc90":2}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{
		AppId:          7,
		Network:        "edge-net",
		Step:           120,
		FastedgeMetric: "min",
		AppName:        "App7",
	}
	frames, err := cli.QueryFastEdge(context.Background(), qm, feTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: %d", len(frames))
	}
	if frames[0].Name != "App7 min" {
		t.Fatalf("name=%q", frames[0].Name)
	}
}

func TestQueryFastEdge_MetricPerc90(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stats":[{"time":"2026-01-01T00:00:00Z","avg":1,"min":0,"max":0,"median":0,"perc75":0,"perc90":9.5}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{FastedgeMetric: "perc90", Step: 60}
	frames, err := cli.QueryFastEdge(context.Background(), qm, feTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatal("expected one frame")
	}
}

func TestQueryFastEdge_EmptyStats(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stats":[]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	frames, err := cli.QueryFastEdge(context.Background(), &core.QueryModel{Step: 60}, feTR())
	if err != nil {
		t.Fatal(err)
	}
	if frames != nil && len(frames) != 0 {
		t.Fatalf("expected nil/empty frames, got %d", len(frames))
	}
}

func TestQueryFastEdge_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	_, err := cli.QueryFastEdge(context.Background(), &core.QueryModel{Step: 60}, feTR())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQueryFastEdge_InvalidTimeSkipped(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/fastedge/v1/stats/app_duration", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stats":[{"time":"not-rfc3339","avg":1,"min":0,"max":0,"median":0,"perc75":0,"perc90":0},{"time":"2026-01-01T01:00:00Z","avg":2,"min":0,"max":0,"median":0,"perc75":0,"perc90":0}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	frames, err := cli.QueryFastEdge(context.Background(), &core.QueryModel{Step: 60}, feTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 || frames[0].Rows() != 1 {
		t.Fatalf("expected one row after skip bad time, rows=%d", frames[0].Rows())
	}
}
