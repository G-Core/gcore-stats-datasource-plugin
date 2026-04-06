package edgenetwork_test

import (
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/edgenetwork"
)

func TestParseWaapSeries_AlternateKeys(t *testing.T) {
	raw := map[string][]map[string]interface{}{
		"total_requests": {
			{"datetime": "2026-01-01T00:00:00Z", "count": float64(3)},
		},
	}
	got := edgenetwork.ParseWaapSeries(raw)
	points := got["total_requests"]
	if len(points) != 1 {
		t.Fatalf("len=%d", len(points))
	}
	if points[0].DateTime != "2026-01-01T00:00:00Z" || points[0].Value != 3 {
		t.Fatalf("point: %+v", points[0])
	}
}

func TestParseWaapSeries_EmptySlice(t *testing.T) {
	if got := edgenetwork.ParseWaapSeries(nil); len(got) != 0 {
		t.Fatalf("expected empty map")
	}
}

func TestParseWaapSeries_SkipsBadValues(t *testing.T) {
	raw := map[string][]map[string]interface{}{
		"m": {
			{"time": "2026-01-01T00:00:00Z", "value": "not-a-number"},
		},
	}
	got := edgenetwork.ParseWaapSeries(raw)
	if got["m"][0].Value != 0 {
		t.Fatalf("expected 0 value, got %v", got["m"][0].Value)
	}
}

func TestParseWaapSeries_DateFieldPriority(t *testing.T) {
	raw := map[string][]map[string]interface{}{
		"m": {
			{"date_time": "a", "timestamp": "b"},
		},
	}
	got := edgenetwork.ParseWaapSeries(raw)
	if got["m"][0].DateTime != "a" {
		t.Fatalf("got %q", got["m"][0].DateTime)
	}
}
