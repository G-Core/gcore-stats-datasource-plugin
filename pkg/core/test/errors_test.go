package core_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
)

func TestHandleAPIError_OK(t *testing.T) {
	if err := core.HandleAPIError(http.StatusOK, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := core.HandleAPIError(http.StatusNoContent, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestHandleAPIError_KnownStatuses(t *testing.T) {
	cases := []struct {
		code int
		body string
		want string
	}{
		{http.StatusBadRequest, `{"error":"bad"}`, "bad request"},
		{http.StatusUnauthorized, `{}`, "authentication"},
		{http.StatusForbidden, `{}`, "forbidden"},
		{http.StatusNotFound, `{}`, "not found"},
		{http.StatusTooManyRequests, `{}`, "many requests"},
		{http.StatusInternalServerError, `{"detail":"oops"}`, "server error"},
		{http.StatusTeapot, `plain`, "gcore api error"},
	}
	for _, tc := range cases {
		err := core.HandleAPIError(tc.code, []byte(tc.body))
		if err == nil {
			t.Fatalf("code %d: expected error", tc.code)
		}
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
			t.Fatalf("code %d: got %q want substring %q", tc.code, err.Error(), tc.want)
		}
	}
}

func TestHandleAPIError_JSONMessageFallback(t *testing.T) {
	err := core.HandleAPIError(http.StatusBadRequest, []byte(`{"message":"msg"}`))
	if err == nil || !strings.Contains(err.Error(), "msg") {
		t.Fatalf("got %v", err)
	}
}
