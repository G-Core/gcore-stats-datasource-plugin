package core_test

import (
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
)

func TestParseRFC3339(t *testing.T) {
	if _, ok := core.ParseRFC3339("2026-01-01T00:00:00Z"); !ok {
		t.Fatal("expected valid RFC3339 timestamp to parse")
	}
	if _, ok := core.ParseRFC3339("not-a-time"); ok {
		t.Fatal("expected invalid timestamp to fail parsing")
	}
}
