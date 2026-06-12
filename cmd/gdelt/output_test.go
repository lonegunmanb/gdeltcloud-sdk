package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what was
// written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

func TestEventsEmitsFullEnvelope(t *testing.T) {
	var gotQuery string
	body := `{"success":true,"data":[{"id":"e1","actors":[{"name":"x"}],"top_articles":[{"url":"u"}]}],"pagination":{"limit":25,"cursor":null,"next_cursor":"25"}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	t.Setenv("GDELT_API_KEY", "gdelt_sk_test")
	os.Unsetenv("GDELT_BASE_URL")

	out := captureStdout(t, func() {
		if code := cmdEvents([]string{"--base-url", srv.URL, "--country", "FRA", "--limit", "25", "--cursor", "0"}); code != 0 {
			t.Fatalf("exit = %d, want 0", code)
		}
	})

	// The cursor flag must reach the API.
	if !strings.Contains(gotQuery, "cursor=0") {
		t.Errorf("query %q missing cursor=0", gotQuery)
	}

	// Output must be the documented envelope, parseable as an object whose
	// `data` is an array and which carries `pagination.next_cursor`.
	var env struct {
		Data       []map[string]json.RawMessage `json:"data"`
		Pagination struct {
			NextCursor json.RawMessage `json:"next_cursor"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("output is not a JSON envelope object: %v\noutput: %s", err, out)
	}
	if len(env.Data) != 1 {
		t.Fatalf("data array length = %d, want 1; output: %s", len(env.Data), out)
	}
	if string(env.Pagination.NextCursor) != `"25"` {
		t.Errorf("pagination.next_cursor = %s, want \"25\"", env.Pagination.NextCursor)
	}
	if _, ok := env.Data[0]["actors"]; !ok {
		t.Errorf("data record dropped actors; output: %s", out)
	}
	if _, ok := env.Data[0]["top_articles"]; !ok {
		t.Errorf("data record dropped top_articles; output: %s", out)
	}
}

func TestEventsCompactEmitsSingleLine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"e1"}],"pagination":{"next_cursor":null}}`))
	}))
	defer srv.Close()

	t.Setenv("GDELT_API_KEY", "gdelt_sk_test")
	os.Unsetenv("GDELT_BASE_URL")

	out := captureStdout(t, func() {
		if code := cmdEvents([]string{"--base-url", srv.URL, "--compact"}); code != 0 {
			t.Fatalf("exit = %d, want 0", code)
		}
	})
	trimmed := strings.TrimRight(out, "\n")
	if strings.Contains(trimmed, "\n") {
		t.Errorf("compact output should be single-line, got:\n%s", out)
	}
	if !json.Valid([]byte(trimmed)) {
		t.Errorf("compact output is not valid JSON: %s", trimmed)
	}
}
