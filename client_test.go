package gdeltcloud

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("expected error for empty API key")
	}
	if _, err := NewClient("   "); err == nil {
		t.Fatal("expected error for blank API key")
	}
	if _, err := NewClient("gdelt_sk_test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEventsRequestAndDecode(t *testing.T) {
	var gotPath string
	var gotQuery string
	var gotAuth string
	var gotUA string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"id":"e1","title":"Strike","event_date":"2026-05-01",
			 "geo":{"latitude":12.5,"longitude":43.2,"country":"YEM"},
			 "metrics":{"magnitude":3.2,"goldstein_scale":-7.5,"fatalities":2}}
		]}`))
	}))
	defer srv.Close()

	c, err := NewClient("gdelt_sk_test", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	events, err := c.Events(context.Background(), EventsParams{
		Country:          []string{"YEM", "SAU"},
		StartDate:        "2026-04-21",
		EndDate:          "2026-05-21",
		Limit:            50,
		IncludeImages:    false,
		HasIncludeImages: true,
	})
	if err != nil {
		t.Fatalf("Events returned error: %v", err)
	}

	if gotPath != "/api/v2/events" {
		t.Errorf("path = %q, want /api/v2/events", gotPath)
	}
	wantAuth := "Bearer " + "gdelt_sk_test"
	if gotAuth != wantAuth {
		t.Errorf("auth header = %q, want %q", gotAuth, wantAuth)
	}
	if gotUA == "" {
		t.Error("expected User-Agent header to be set")
	}
	wantContains := []string{"country=YEM%2CSAU", "start_date=2026-04-21", "end_date=2026-05-21", "limit=50", "include_images=false"}
	for _, want := range wantContains {
		if !containsSub(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	e := events[0]
	if e.ID != "e1" || e.Title != "Strike" {
		t.Errorf("unexpected event: %+v", e)
	}
	if e.Geo == nil || e.Geo.Country != "YEM" || e.Geo.Latitude == nil || *e.Geo.Latitude != 12.5 {
		t.Errorf("unexpected geo: %+v", e.Geo)
	}
	if e.Metrics == nil || e.Metrics.GoldsteinScale == nil || *e.Metrics.GoldsteinScale != -7.5 {
		t.Errorf("unexpected metrics: %+v", e.Metrics)
	}
}

func TestIncludeImagesOmittedWhenNotSet(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer srv.Close()

	c, _ := NewClient("k", WithBaseURL(srv.URL))
	if _, err := c.Events(context.Background(), EventsParams{StartDate: "2026-01-01"}); err != nil {
		t.Fatal(err)
	}
	if containsSub(gotQuery, "include_images") {
		t.Errorf("include_images should be omitted, query = %q", gotQuery)
	}
}

func TestStoriesAndEntitiesAndEnergyPaths(t *testing.T) {
	paths := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths[r.URL.Path] = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	if _, err := c.Stories(context.Background(), StoriesParams{ArticleCountMin: 4, Limit: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Entities(context.Background(), EntitiesParams{Search: "Houthi"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.EnergyAssets(context.Background(), EnergyAssetsParams{Bbox: "11.5,42.5,13.5,44.5", Tracker: []string{"oil_gas_plants", "lng_terminals"}}); err != nil {
		t.Fatal(err)
	}

	if q, ok := paths["/api/v2/stories"]; !ok || !containsSub(q, "article_count_min=4") {
		t.Errorf("stories not called correctly: %q", q)
	}
	if q, ok := paths["/api/v2/entities"]; !ok || !containsSub(q, "search=Houthi") {
		t.Errorf("entities not called correctly: %q", q)
	}
	q, ok := paths["/api/v2/energy/assets"]
	if !ok || !containsSub(q, "tracker=oil_gas_plants%2Clng_terminals") {
		t.Errorf("energy assets not called correctly: %q", q)
	}
}

func TestErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":false,"error":"invalid date range"}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	_, err := c.Events(context.Background(), EventsParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message != "invalid date range" {
		t.Errorf("unexpected message: %q", apiErr.Message)
	}
}

func TestHTTPErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"unauthorized"}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	_, err := c.Events(context.Background(), EventsParams{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", apiErr.StatusCode)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html>502 Bad Gateway</html>`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	_, err := c.Events(context.Background(), EventsParams{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", apiErr.StatusCode)
	}
}

func TestWithHTTPClientAndTimeout(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	c, err := NewClient("k", WithHTTPClient(hc), WithUserAgent("custom-agent"))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient != hc {
		t.Error("custom HTTP client not used")
	}
	if c.userAgent != "custom-agent" {
		t.Error("custom user agent not used")
	}
}

func TestMissingSuccessFieldTreatedAsOK(t *testing.T) {
	// Some responses may omit "success"; absence should not be treated as error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"x"}]}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))
	events, err := c.Events(context.Background(), EventsParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 || events[0].ID != "x" {
		t.Errorf("unexpected events: %+v", events)
	}
}

func containsSub(s, sub string) bool {
	return strings.Contains(s, sub)
}
