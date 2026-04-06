package datasource_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type recordingResourceSender struct {
	resp *backend.CallResourceResponse
}

func (r *recordingResourceSender) Send(resp *backend.CallResourceResponse) error {
	r.resp = resp
	return nil
}

func TestGCDataSource_CallResource_StripsGrafanaAccessControlQuery(t *testing.T) {
	t.Parallel()

	var upstream string
	mux := http.NewServeMux()
	mux.HandleFunc("/iam/users/me", func(w http.ResponseWriter, r *http.Request) {
		upstream = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "secret")
	ds.Client = srv.Client()

	sender := &recordingResourceSender{}
	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path:   "iam/users/me",
		Method: http.MethodGet,
		URL:    "iam/users/me?accesscontrol=true",
	}, sender)
	if err != nil {
		t.Fatal(err)
	}
	if upstream != "/iam/users/me" {
		t.Fatalf("upstream URL %q, want no Grafana-only query", upstream)
	}
	if sender.resp == nil || sender.resp.Status != 200 {
		t.Fatalf("response: %+v", sender.resp)
	}
}

func TestGCDataSource_CallResource_KeepsRealQueryParams(t *testing.T) {
	t.Parallel()

	var upstream string
	mux := http.NewServeMux()
	mux.HandleFunc("/dns/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		upstream = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"zones":[]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "secret")
	ds.Client = srv.Client()

	sender := &recordingResourceSender{}
	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path:   "dns/zones",
		Method: http.MethodGet,
		URL:    "dns/zones?accesscontrol=true&limit=10&offset=0",
	}, sender)
	if err != nil {
		t.Fatal(err)
	}
	if upstream != "/dns/v2/zones?limit=10&offset=0" && upstream != "/dns/v2/zones?offset=0&limit=10" {
		t.Fatalf("upstream URL %q, want limit/offset only", upstream)
	}
}

func TestGCDataSource_CallResource_PathFromURLWhenPathEmpty(t *testing.T) {
	t.Parallel()

	var upstream string
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		upstream = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"legacy"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ds := datasource.NewDataSource(srv.URL, "secret")
	ds.Client = srv.Client()

	sender := &recordingResourceSender{}
	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path:   "",
		Method: http.MethodGet,
		URL:    "/users/me?accesscontrol=true",
	}, sender)
	if err != nil {
		t.Fatal(err)
	}
	if upstream != "/users/me" {
		t.Fatalf("upstream URL %q", upstream)
	}
}
