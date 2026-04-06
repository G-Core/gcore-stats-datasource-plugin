package core_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
)

func TestApplyJSONAuthHeaders_APIKeyRaw(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example", nil)
	core.ApplyJSONAuthHeaders(req, "abc")
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}
	if got := req.Header.Get("Authorization"); got != "APIKey abc" {
		t.Fatalf("expected Authorization %q, got %q", "APIKey abc", got)
	}
}

func TestApplyJSONAuthHeaders_PassthroughBearer(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example", nil)
	core.ApplyJSONAuthHeaders(req, "Bearer token")
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("expected Authorization %q, got %q", "Bearer token", got)
	}
}

func TestDoRequest_ResponseTooLarge(t *testing.T) {
	old := core.DefaultMaxResponseBodyBytes
	core.DefaultMaxResponseBodyBytes = 64
	t.Cleanup(func() { core.DefaultMaxResponseBodyBytes = old })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("a"), int(core.DefaultMaxResponseBodyBytes)+1))
	}))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	_, _, err := core.DoRequest(srv.Client(), req, nil)
	if !errors.Is(err, core.ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestDoJSONRequest_UnmarshalOnAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"x":1}`)
	}))
	t.Cleanup(srv.Close)

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	var out struct {
		X int `json:"x"`
	}
	raw, err := core.DoJSONRequest(srv.Client(), req, &out, nil, core.HandleAPIError, http.StatusOK)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if out.X != 1 {
		t.Fatalf("expected out.X=1, got %d", out.X)
	}
	if string(raw) != `{"x":1}` {
		t.Fatalf("expected raw body, got %q", string(raw))
	}
}

func TestDoJSON_BuildsAndExecutesRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	var out struct {
		OK bool `json:"ok"`
	}
	_, err := core.DoJSON(context.Background(), srv.Client(), http.MethodGet, srv.URL, nil, &out, nil, core.HandleAPIError, http.StatusOK)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true")
	}
}
