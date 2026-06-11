package gdeltcloud

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestByIDEndpoints(t *testing.T) {
	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/events/conflict_1":
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"conflict_1","title":"Strike","summary":"s","geo":{"admin1":"Ile-de-France","location":"Paris"}}}`))
		case "/api/v2/stories/story_1":
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"story_1","title":"Cluster"}}`))
		case "/api/v2/entities/person:Example Person":
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"person:Example Person","name":"Example Person","type":"person"}}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))
	ctx := context.Background()

	ev, err := c.Event(ctx, "conflict_1")
	if err != nil {
		t.Fatalf("Event: %v", err)
	}
	if ev.ID != "conflict_1" || ev.Summary != "s" {
		t.Errorf("unexpected event: %+v", ev)
	}
	if ev.Geo == nil || ev.Geo.Admin1 != "Ile-de-France" || ev.Geo.Location != "Paris" {
		t.Errorf("unexpected geo: %+v", ev.Geo)
	}

	st, err := c.Story(ctx, "story_1")
	if err != nil {
		t.Fatalf("Story: %v", err)
	}
	if st.ID != "story_1" {
		t.Errorf("unexpected story: %+v", st)
	}

	en, err := c.Entity(ctx, "person:Example Person")
	if err != nil {
		t.Fatalf("Entity: %v", err)
	}
	if en.Type != "person" || en.Name != "Example Person" {
		t.Errorf("unexpected entity: %+v", en)
	}

	want := []string{"/api/v2/events/conflict_1", "/api/v2/stories/story_1", "/api/v2/entities/person:Example Person"}
	for i, p := range want {
		if i >= len(gotPaths) || gotPaths[i] != p {
			t.Errorf("path[%d] = %q, want %q", i, gotPaths, p)
		}
	}
}

func TestByIDRequiresID(t *testing.T) {
	c, _ := NewClient("k")
	ctx := context.Background()
	if _, err := c.Event(ctx, ""); err == nil {
		t.Error("Event(\"\") should error")
	}
	if _, err := c.Story(ctx, "  "); err == nil {
		t.Error("Story(blank) should error")
	}
	if _, err := c.Entity(ctx, ""); err == nil {
		t.Error("Entity(\"\") should error")
	}
	if _, err := c.GeoAdmin1(ctx, ""); err == nil {
		t.Error("GeoAdmin1(\"\") should error")
	}
}

func TestSummaryEndpoints(t *testing.T) {
	paths := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths[r.URL.Path] = r.URL.RawQuery
		switch r.URL.Path {
		case "/api/v2/events/summary":
			_, _ = w.Write([]byte(`{"success":true,"group_by":"country","data":[
				{"key":"France","group_by":"country","event_count":12,"fatalities":6,
				 "metrics":{"significance":{"avg":0.42,"max":0.91}}}
			]}`))
		case "/api/v2/stories/summary":
			_, _ = w.Write([]byte(`{"success":true,"group_by":"date","data":[
				{"key":"2026-04-17","group_by":"date","story_count":5,"linked_event_count":2}
			]}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))
	ctx := context.Background()

	buckets, err := c.EventsSummary(ctx, EventsSummaryParams{
		GroupBy:          "country",
		Region:           "Middle East",
		Category:         []string{"Protests", "INFRASTRUCTURE"},
		StartDate:        "2026-04-01",
		EndDate:          "2026-04-17",
		HasFatalities:    true,
		HasHasFatalities: true,
	})
	if err != nil {
		t.Fatalf("EventsSummary: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Key != "France" || buckets[0].EventCount == nil || *buckets[0].EventCount != 12 {
		t.Errorf("unexpected event buckets: %+v", buckets)
	}
	if len(buckets[0].Metrics) == 0 {
		t.Error("expected raw metrics to be preserved")
	}
	q := paths["/api/v2/events/summary"]
	for _, want := range []string{"group_by=country", "region=Middle+East", "category=Protests%2CINFRASTRUCTURE", "date_start=2026-04-01", "has_fatalities=true"} {
		if !containsSub(q, want) {
			t.Errorf("events summary query %q missing %q", q, want)
		}
	}

	sBuckets, err := c.StoriesSummary(ctx, StoriesSummaryParams{
		GroupBy:         "date",
		Continent:       "Asia",
		ArticleCountMin: 2,
	})
	if err != nil {
		t.Fatalf("StoriesSummary: %v", err)
	}
	if len(sBuckets) != 1 || sBuckets[0].StoryCount == nil || *sBuckets[0].StoryCount != 5 {
		t.Errorf("unexpected story buckets: %+v", sBuckets)
	}
	q = paths["/api/v2/stories/summary"]
	for _, want := range []string{"group_by=date", "continent=Asia", "article_count_min=2"} {
		if !containsSub(q, want) {
			t.Errorf("stories summary query %q missing %q", q, want)
		}
	}
}

func TestGeoAdmin1(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.URL.Path != "/api/v2/geo/admin1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"country":"France","admin1":["Bretagne","Ile-de-France"],"source":"gdelt_cloud.events"}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	res, err := c.GeoAdmin1(context.Background(), "France")
	if err != nil {
		t.Fatalf("GeoAdmin1: %v", err)
	}
	if res.Country != "France" || res.Source != "gdelt_cloud.events" {
		t.Errorf("unexpected admin1 envelope: %+v", res)
	}
	if len(res.Admin1) != 2 || res.Admin1[0] != "Bretagne" {
		t.Errorf("unexpected admin1 list: %+v", res.Admin1)
	}
	if !containsSub(gotQuery, "country=France") {
		t.Errorf("query %q missing country=France", gotQuery)
	}
}

func TestSummaryErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"error":"group_by required","code":"INVALID_GROUP_BY"}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	_, err := c.EventsSummary(context.Background(), EventsSummaryParams{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.Message != "group_by required" {
		t.Errorf("unexpected message: %q", apiErr.Message)
	}
}
