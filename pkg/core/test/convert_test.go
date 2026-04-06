package core_test

import (
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/core"
)

func TestToString(t *testing.T) {
	if got := core.ToString(42); got != "42" {
		t.Fatalf("got %q", got)
	}
	if got := core.ToString("x"); got != "x" {
		t.Fatalf("got %q", got)
	}
}

func TestToFloat64(t *testing.T) {
	n, ok := core.ToFloat64(float64(1.5))
	if !ok || n != 1.5 {
		t.Fatalf("float64: ok=%v n=%v", ok, n)
	}
	n, ok = core.ToFloat64(int(7))
	if !ok || n != 7 {
		t.Fatalf("int: ok=%v n=%v", ok, n)
	}
	_, ok = core.ToFloat64("nope")
	if ok {
		t.Fatal("expected false for string")
	}
}
