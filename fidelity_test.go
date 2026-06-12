package gdeltcloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fullEventEnvelope mirrors the documented v2 events response: a success
// envelope with a pagination block and an Event card carrying the fields that
// the typed Event struct historically dropped.
const fullEventEnvelope = `{
  "success": true,
  "data": [
    {
      "id": "conflict_20260417_example",
      "url": "https://gdeltcloud.com/story/example",
      "primary_story_url": "https://gdeltcloud.com/story/example",
      "family": "conflict",
      "title": "Example event title",
      "summary": "Short summary.",
      "event_date": "2026-04-17",
      "category": "Protests",
      "subcategory": "Peaceful protest",
      "domain": "CONFLICT",
      "event_code": "ACLED-123",
      "geo": {"country": "France", "latitude": 48.8566, "longitude": 2.3522},
      "geo_context": {"location_country": "France", "actor_origin_countries": ["France"]},
      "actors": [{"name": "Protesters", "country": "France", "role": "actor1"}],
      "metrics": {"significance": 0.72, "goldstein_scale": -2, "confidence": 0.91, "article_count": 4},
      "has_fatalities": true,
      "fatalities": 3,
      "civilian_targeting": false,
      "civilian_targeting_label": null,
      "story_refs": [{"id": "story_1", "url": "https://gdeltcloud.com/story/example", "title": "Example story", "story_date": "2026-04-17", "article_count": 4}],
      "entity_refs": [{"id": "Paris", "name": "Paris", "type": "LOCATION", "wikipedia_url": "https://en.wikipedia.org/wiki/Paris"}],
      "top_articles": [{"url": "https://example.com/article", "title": "Example article", "rank": 1}]
    }
  ],
  "pagination": {"limit": 25, "cursor": null, "next_cursor": "25"}
}`

func TestEventsRawPreservesEnvelopeAndFields(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(fullEventEnvelope))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	raw, err := c.EventsRaw(context.Background(), EventsParams{Country: []string{"FRA"}, Limit: 25, Cursor: "0"})
	if err != nil {
		t.Fatalf("EventsRaw: %v", err)
	}

	// The cursor parameter must be sent to the API.
	if !strings.Contains(gotQuery, "cursor=0") {
		t.Errorf("query %q missing cursor=0", gotQuery)
	}

	// The raw body must preserve the envelope: success, data and pagination,
	// plus every documented record field that the typed struct used to drop.
	var env struct {
		Success    *bool             `json:"success"`
		Data       []json.RawMessage `json:"data"`
		Pagination struct {
			Limit      *int            `json:"limit"`
			NextCursor json.RawMessage `json:"next_cursor"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal raw envelope: %v", err)
	}
	if env.Success == nil || !*env.Success {
		t.Error("success flag not preserved")
	}
	if env.Pagination.Limit == nil || *env.Pagination.Limit != 25 {
		t.Errorf("pagination.limit not preserved: %+v", env.Pagination)
	}
	if string(env.Pagination.NextCursor) != `"25"` {
		t.Errorf("pagination.next_cursor = %s, want \"25\"", env.Pagination.NextCursor)
	}
	if len(env.Data) != 1 {
		t.Fatalf("got %d records, want 1", len(env.Data))
	}
	for _, field := range []string{
		"primary_story_url", "event_code", "geo_context", "actors",
		"has_fatalities", "civilian_targeting", "story_refs", "entity_refs",
		"top_articles", "confidence", "article_count",
	} {
		if !strings.Contains(string(env.Data[0]), field) {
			t.Errorf("raw record dropped field %q", field)
		}
	}
}

func TestEventDecodesFullCard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fullEventEnvelope))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	events, err := c.Events(context.Background(), EventsParams{})
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	e := events[0]
	if e.PrimaryStoryURL == "" || e.EventCode != "ACLED-123" {
		t.Errorf("primary_story_url/event_code not decoded: %+v", e)
	}
	if e.GeoContext == nil || e.GeoContext.LocationCountry != "France" || len(e.GeoContext.ActorOriginCountries) != 1 {
		t.Errorf("geo_context not decoded: %+v", e.GeoContext)
	}
	if len(e.Actors) != 1 || e.Actors[0].Name != "Protesters" || e.Actors[0].Role != "actor1" {
		t.Errorf("actors not decoded: %+v", e.Actors)
	}
	if e.HasFatalities == nil || !*e.HasFatalities {
		t.Errorf("has_fatalities not decoded: %+v", e.HasFatalities)
	}
	if e.Fatalities == nil || *e.Fatalities != 3 {
		t.Errorf("fatalities not decoded: %+v", e.Fatalities)
	}
	if e.CivilianTargeting == nil || *e.CivilianTargeting {
		t.Errorf("civilian_targeting not decoded: %+v", e.CivilianTargeting)
	}
	if len(e.StoryRefs) != 1 || e.StoryRefs[0].ID != "story_1" || e.StoryRefs[0].ArticleCount == nil {
		t.Errorf("story_refs not decoded: %+v", e.StoryRefs)
	}
	if len(e.EntityRefs) != 1 || e.EntityRefs[0].Type != "LOCATION" {
		t.Errorf("entity_refs not decoded: %+v", e.EntityRefs)
	}
	if len(e.TopArticles) != 1 || e.TopArticles[0].Rank == nil || *e.TopArticles[0].Rank != 1 {
		t.Errorf("top_articles not decoded: %+v", e.TopArticles)
	}
	if e.Metrics == nil || e.Metrics.Confidence == nil || *e.Metrics.Confidence != 0.91 {
		t.Errorf("metrics.confidence not decoded: %+v", e.Metrics)
	}
	if e.Metrics.ArticleCount == nil || *e.Metrics.ArticleCount != 4 {
		t.Errorf("metrics.article_count not decoded: %+v", e.Metrics)
	}
}

func TestStoryDecodesFullCard(t *testing.T) {
	const body = `{"success":true,"data":{
		"id":"story_1","title":"Example","story_date":"2026-04-17",
		"geo_context":{"location_country":"Japan","actor_origin_countries":["Japan","United States"]},
		"metrics":{"significance":0.82,"article_count":12,"linked_event_count":2,"max_linked_event_significance":0.76},
		"has_events":true,"has_fatalities":false,"fatalities":0,
		"linked_events":[{"id":"cameoplus_1","title":"Linked"}],
		"entity_refs":[{"id":"Tokyo","name":"Tokyo","type":"LOCATION"}],
		"top_articles":[{"url":"https://x","title":"a","domain":"x.com","rank":1}]
	}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	st, err := c.Story(context.Background(), "story_1")
	if err != nil {
		t.Fatalf("Story: %v", err)
	}
	if st.GeoContext == nil || len(st.GeoContext.ActorOriginCountries) != 2 {
		t.Errorf("geo_context not decoded: %+v", st.GeoContext)
	}
	if st.HasEvents == nil || !*st.HasEvents {
		t.Errorf("has_events not decoded: %+v", st.HasEvents)
	}
	if st.Metrics == nil || st.Metrics.MaxLinkedEventSignificance == nil || *st.Metrics.MaxLinkedEventSignificance != 0.76 {
		t.Errorf("metrics.max_linked_event_significance not decoded: %+v", st.Metrics)
	}
	if len(st.LinkedEvents) != 1 || st.LinkedEvents[0].ID != "cameoplus_1" {
		t.Errorf("linked_events not decoded: %+v", st.LinkedEvents)
	}
	if len(st.EntityRefs) != 1 || st.EntityRefs[0].Name != "Tokyo" {
		t.Errorf("entity_refs not decoded: %+v", st.EntityRefs)
	}
}

func TestSummaryRawPreservesGroupBy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"group_by":"country","data":[{"key":"France","event_count":1}]}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	raw, err := c.EventsSummaryRaw(context.Background(), EventsSummaryParams{GroupBy: "country"})
	if err != nil {
		t.Fatalf("EventsSummaryRaw: %v", err)
	}
	if !strings.Contains(string(raw), `"group_by":"country"`) {
		t.Errorf("top-level group_by not preserved: %s", raw)
	}
}

func TestRawSurfacesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"error":"bad"}`))
	}))
	defer srv.Close()
	c, _ := NewClient("k", WithBaseURL(srv.URL))

	if _, err := c.EventsRaw(context.Background(), EventsParams{}); err == nil {
		t.Fatal("expected error from EventsRaw")
	} else if _, ok := err.(*APIError); !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
}
