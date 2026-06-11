package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"bogus"}); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestRunVersion(t *testing.T) {
	if code := run([]string{"version"}); code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
}

func TestEventsRequiresAPIKey(t *testing.T) {
	t.Setenv("GDELT_API_KEY", "")
	t.Setenv("GDELT_BASE_URL", "")
	if code := cmdEvents([]string{"--country", "YEM"}); code != 2 {
		t.Errorf("exit = %d, want 2 when API key missing", code)
	}
}

func TestEventsAgainstMockServer(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"e1"}]}`))
	}))
	defer srv.Close()

	t.Setenv("GDELT_API_KEY", "gdelt_sk_test")
	code := cmdEvents([]string{
		"--base-url", srv.URL,
		"--country", "YEM,SAU",
		"--start", "2026-04-21",
		"--end", "2026-05-21",
		"--limit", "5",
	})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if gotPath != "/api/v2/events" {
		t.Errorf("path = %q", gotPath)
	}
	for _, want := range []string{"country=YEM%2CSAU", "start_date=2026-04-21", "limit=5"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
}

func TestSplitCSV(t *testing.T) {
	cases := map[string][]string{
		"":             nil,
		"  ":           nil,
		"YEM":          {"YEM"},
		"YEM,SAU":      {"YEM", "SAU"},
		" YEM , SAU ,": {"YEM", "SAU"},
	}
	for in, want := range cases {
		got := splitCSV(in)
		if len(got) != len(want) {
			t.Errorf("splitCSV(%q) = %v, want %v", in, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitCSV(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestApiKeyFromEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer srv.Close()

	t.Setenv("GDELT_API_KEY", "gdelt_sk_env")
	os.Unsetenv("GDELT_BASE_URL")
	if code := cmdStories([]string{"--base-url", srv.URL}); code != 0 {
		t.Errorf("exit = %d, want 0 using env API key", code)
	}
}
