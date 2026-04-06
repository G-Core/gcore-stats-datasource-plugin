package datasource_test

import (
	"fmt"
	"testing"

	"github.com/G-Core/gcore-stats-datasource-plugin/pkg/datasource"
)

func TestValidateAPIBaseURL(t *testing.T) {
	t.Parallel()

	valid := []string{
		"",
		"   ",
		"api.gcore.com",
		"https://api.gcore.com",
		"https://api.gcore.com/",
		"http://api.gcore.com",
		"api.staging.gcore.com",
		"API.GCORE.COM",
		"api.gcore.com:443",
		"api.custom.example.com",
		"https://api.evil.com",
	}
	for i, raw := range valid {
		raw := raw
		t.Run(fmt.Sprintf("valid_%d_%q", i, raw), func(t *testing.T) {
			t.Parallel()
			if err := datasource.ValidateAPIBaseURL(raw); err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}

	invalid := []string{
		"cdn.gcore.com",
		"https://evil.com",
		"api.gcore.com/foo",
		"https://api.gcore.com/path",
		"api.gcore.org",
		"notapi.gcore.com",
		"https://user:pass@api.gcore.com",
		"https://api.gcore.com?q=1",
		"https://api.gcore.com#frag",
		"ftp://api.gcore.com",
	}
	for i, raw := range invalid {
		raw := raw
		t.Run(fmt.Sprintf("invalid_%d_%q", i, raw), func(t *testing.T) {
			t.Parallel()
			if err := datasource.ValidateAPIBaseURL(raw); err == nil {
				t.Fatal("expected invalid")
			}
		})
	}
}
