package edgenetwork_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func dnsTR() backend.TimeRange {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return backend.TimeRange{From: from, To: from.Add(time.Hour)}
}

func TestQueryDNS_SingleExplicitZone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/my.example/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("granularity") != "5m" {
			t.Errorf("granularity=%q", r.URL.Query().Get("granularity"))
		}
		if r.URL.Query().Get("record_type") != "" {
			t.Errorf("unexpected record_type for ALL")
		}
		_, _ = w.Write([]byte(`{"requests":{"1704067200":12}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, APIKey: "k", HTTP: srv.Client()}
	qm := &core.QueryModel{
		Zone:           "my.example",
		RecordType:     "ALL",
		DnsGranularity: core.SelectableValue{Value: "5m", Label: "5m"},
	}
	frames, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: %d", len(frames))
	}
	if frames[0].Fields[1].Config == nil || frames[0].Fields[1].Config.DisplayName != "Requests (my.example)" {
		t.Fatalf("display: %#v", frames[0].Fields[1].Config)
	}
}

func TestQueryDNS_RecordTypeInQuery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/z/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("record_type") != "A" {
			t.Errorf("record_type=%q", r.URL.Query().Get("record_type"))
		}
		_, _ = w.Write([]byte(`{"requests":{}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{
		Zone:           "z",
		RecordType:     "A",
		DnsGranularity: core.SelectableValue{Value: "1h"},
	}
	_, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryDNS_AllZones_Aggregates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"zones":[{"name":"a"},{"name":"b"}],"total_amount":2}`))
	})
	mux.HandleFunc("/dns/v2/zones/a/statistics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"requests":{"1704067200":1}}`))
	})
	mux.HandleFunc("/dns/v2/zones/b/statistics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"requests":{"1704067200":2}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{
		Zone:           "all",
		RecordType:     "ALL",
		DnsGranularity: core.SelectableValue{Value: "5m"},
	}
	frames, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: %d", len(frames))
	}
	if frames[0].Rows() != 1 {
		t.Fatalf("rows: %d", frames[0].Rows())
	}
	v, err := frames[0].FloatAt(1, 0)
	if err != nil || v != 3 {
		t.Fatalf("aggregated value: %v err=%v", v, err)
	}
}

func TestQueryDNS_ZoneStatsNotFound_Skips(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/only/statistics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{
		Zone:           "only",
		DnsGranularity: core.SelectableValue{Value: "5m"},
	}
	frames, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 || frames[0].Rows() != 0 {
		t.Fatalf("expected empty stats frame")
	}
}

func TestQueryDNS_ZoneListError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{Zone: "all", DnsGranularity: core.SelectableValue{Value: "5m"}}
	_, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQueryDNS_ZoneStatsError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones/bad/statistics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{Zone: "bad", DnsGranularity: core.SelectableValue{Value: "5m"}}
	_, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQueryDNS_NoZonesEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"zones":[],"total_amount":0}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cli := &edgenetwork.Client{RootURL: srv.URL, HTTP: srv.Client()}
	qm := &core.QueryModel{Zone: "all", DnsGranularity: core.SelectableValue{Value: "5m"}}
	frames, err := cli.QueryDNS(context.Background(), qm, dnsTR())
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 {
		t.Fatalf("frames: %d", len(frames))
	}
}
