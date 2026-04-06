package datasource_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestCheckHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iam/users/me" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(datasource.IAMUser{Name: "Alice"})
	}))
	t.Cleanup(srv.Close)

	ds := &datasource.GCDataSource{
		URL:    srv.URL,
		APIKey: "abc",
		Client: srv.Client(),
	}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if res.Status != backend.HealthStatusOk {
		t.Fatalf("expected OK, got %v (%s)", res.Status, res.Message)
	}
}

func TestCheckHealth_ErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	t.Cleanup(srv.Close)

	ds := &datasource.GCDataSource{
		URL:    srv.URL,
		APIKey: "bad",
		Client: srv.Client(),
	}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if res.Status != backend.HealthStatusError {
		t.Fatalf("expected ERROR, got %v", res.Status)
	}
	if res.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}

func TestCheckHealth_HTTPTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)

	client := srv.Client()
	client.Timeout = 10 * time.Millisecond

	ds := &datasource.GCDataSource{
		URL:    srv.URL,
		APIKey: "abc",
		Client: client,
	}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if res.Status != backend.HealthStatusError {
		t.Fatalf("expected ERROR, got %v", res.Status)
	}
	if res.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}
