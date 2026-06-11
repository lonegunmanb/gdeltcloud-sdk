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
	for _, want := range []string{"country=YEM%2CSAU", "date_start=2026-04-21", "limit=5"} {
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

func TestNewCommandsAgainstMockServer(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v2/events/summary"):
			_, _ = w.Write([]byte(`{"success":true,"group_by":"country","data":[{"key":"France","event_count":1}]}`))
		case strings.HasPrefix(r.URL.Path, "/api/v2/stories/summary"):
			_, _ = w.Write([]byte(`{"success":true,"group_by":"date","data":[{"key":"2026-04-17","story_count":1}]}`))
		case r.URL.Path == "/api/v2/geo/admin1":
			_, _ = w.Write([]byte(`{"success":true,"country":"France","admin1":["Bretagne"]}`))
		default:
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"x"}}`))
		}
	}))
	defer srv.Close()
	t.Setenv("GDELT_API_KEY", "gdelt_sk_test")
	os.Unsetenv("GDELT_BASE_URL")

	cases := []struct {
		name     string
		run      func([]string) int
		args     []string
		wantPath string
		wantQ    string
	}{
		{"event", cmdEvent, []string{"--base-url", "%s", "--id", "conflict_1"}, "/api/v2/events/conflict_1", ""},
		{"story", cmdStory, []string{"--base-url", "%s", "--id", "story_1"}, "/api/v2/stories/story_1", ""},
		{"entity", cmdEntity, []string{"--base-url", "%s", "--id", "person:Foo"}, "/api/v2/entities/person:Foo", ""},
		{"events-summary", cmdEventsSummary, []string{"--base-url", "%s", "--group-by", "country", "--region", "Middle East"}, "/api/v2/events/summary", "group_by=country"},
		{"stories-summary", cmdStoriesSummary, []string{"--base-url", "%s", "--group-by", "date"}, "/api/v2/stories/summary", "group_by=date"},
		{"admin1", cmdAdmin1, []string{"--base-url", "%s", "--country", "France"}, "/api/v2/geo/admin1", "country=France"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := make([]string, len(tc.args))
			for i, a := range tc.args {
				if a == "%s" {
					args[i] = srv.URL
				} else {
					args[i] = a
				}
			}
			if code := tc.run(args); code != 0 {
				t.Fatalf("exit = %d, want 0", code)
			}
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if tc.wantQ != "" && !strings.Contains(gotQuery, tc.wantQ) {
				t.Errorf("query %q missing %q", gotQuery, tc.wantQ)
			}
		})
	}
}

func TestNewCommandsRequiredFlags(t *testing.T) {
	t.Setenv("GDELT_API_KEY", "gdelt_sk_test")
	os.Unsetenv("GDELT_BASE_URL")
	cases := map[string]func([]string) int{
		"event":           cmdEvent,
		"story":           cmdStory,
		"entity":          cmdEntity,
		"events-summary":  cmdEventsSummary,
		"stories-summary": cmdStoriesSummary,
		"admin1":          cmdAdmin1,
	}
	for name, run := range cases {
		if code := run(nil); code == 0 {
			t.Errorf("%s with no required flag: exit = 0, want non-zero", name)
		}
	}
}
