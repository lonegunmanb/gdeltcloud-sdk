package gdeltcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// EventsParams are the query parameters for the events endpoint.
//
// Note: the API caps the events window at 30 days per call; for longer windows
// issue multiple calls and merge the results.
type EventsParams struct {
	// Country filters by source/actor country (ISO-3 codes). Multiple codes
	// are sent as a comma-separated list.
	Country []string
	// StartDate is the inclusive start of the window (ISO date, e.g. "2026-04-21").
	StartDate string
	// EndDate is the inclusive end of the window (ISO date, e.g. "2026-05-21").
	EndDate string
	// Domain optionally filters by CAMEO+ domain (e.g. "INFRASTRUCTURE").
	Domain string
	// Limit caps the number of returned records.
	Limit int
	// Cursor is the pagination cursor: pass the next_cursor value from a prior
	// response's pagination block to fetch the next page.
	Cursor string
	// IncludeImages toggles image enrichment. Use the SetIncludeImages helper
	// or set IncludeImages together with HasIncludeImages.
	IncludeImages    bool
	HasIncludeImages bool
}

func (p EventsParams) values() url.Values {
	v := url.Values{}
	setCSV(v, "country", p.Country)
	setStr(v, "date_start", p.StartDate)
	setStr(v, "date_end", p.EndDate)
	setStr(v, "domain", p.Domain)
	setInt(v, "limit", p.Limit)
	setStr(v, "cursor", p.Cursor)
	setBool(v, "include_images", p.IncludeImages, p.HasIncludeImages)
	return v
}

// Event is a single event returned by the events endpoint. It models the
// documented v2 Event card; callers that need byte-for-byte fidelity to the API
// response (including any field not modeled here) can use EventsRaw / EventRaw.
type Event struct {
	ID                     string        `json:"id,omitempty"`
	EventUID               string        `json:"event_uid,omitempty"`
	Title                  string        `json:"title,omitempty"`
	Summary                string        `json:"summary,omitempty"`
	URL                    string        `json:"url,omitempty"`
	PrimaryStoryURL        string        `json:"primary_story_url,omitempty"`
	EventDate              string        `json:"event_date,omitempty"`
	Date                   string        `json:"date,omitempty"`
	Category               string        `json:"category,omitempty"`
	Subcategory            string        `json:"subcategory,omitempty"`
	Family                 string        `json:"family,omitempty"`
	EventFamily            string        `json:"event_family,omitempty"`
	Domain                 string        `json:"domain,omitempty"`
	EventCode              string        `json:"event_code,omitempty"`
	ImageURL               string        `json:"image_url,omitempty"`
	Geo                    *Geo          `json:"geo,omitempty"`
	GeoContext             *GeoContext   `json:"geo_context,omitempty"`
	Actors                 []Actor       `json:"actors,omitempty"`
	Metrics                *EventMetrics `json:"metrics,omitempty"`
	HasFatalities          *bool         `json:"has_fatalities,omitempty"`
	Fatalities             *int          `json:"fatalities,omitempty"`
	CivilianTargeting      *bool         `json:"civilian_targeting,omitempty"`
	CivilianTargetingLabel string        `json:"civilian_targeting_label,omitempty"`
	StoryRefs              []StoryRef    `json:"story_refs,omitempty"`
	EntityRefs             []EntityRef   `json:"entity_refs,omitempty"`
	TopArticles            []Article     `json:"top_articles,omitempty"`
}

// Events fetches events matching the given parameters.
func (c *Client) Events(ctx context.Context, params EventsParams) ([]Event, error) {
	var out []Event
	if err := c.get(ctx, "/api/v2/events", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EventsRaw fetches events and returns the complete response body verbatim,
// preserving the full success envelope (success, data and pagination) and every
// documented record field. Use it when you need full fidelity to the v2 schema
// or access to the pagination cursor.
func (c *Client) EventsRaw(ctx context.Context, params EventsParams) (json.RawMessage, error) {
	return c.rawBody(ctx, "/api/v2/events", params.values())
}

// Event fetches a single event by its v2 identifier
// (GET /api/v2/events/{event_id}).
func (c *Client) Event(ctx context.Context, id string) (*Event, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gdeltcloud: event id is required")
	}
	var out Event
	if err := c.get(ctx, "/api/v2/events/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EventRaw fetches a single event by id and returns the complete response body
// verbatim. See EventsRaw for why a caller might prefer the raw form.
func (c *Client) EventRaw(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gdeltcloud: event id is required")
	}
	return c.rawBody(ctx, "/api/v2/events/"+url.PathEscape(id), nil)
}

// EventsSummaryParams are the query parameters for the events/summary endpoint.
//
// GroupBy selects the aggregation dimension and is required by the API; valid
// values are "date", "country", "region", "continent", "category" and
// "subcategory". The summary endpoint does not accept the free-text "search"
// parameter.
type EventsSummaryParams struct {
	// GroupBy is the aggregation dimension (required).
	GroupBy string
	// Country filters by country, sent comma-separated.
	Country []string
	// Region expands to a country list on the backend (e.g. "Middle East").
	Region string
	// Continent expands to a country list on the backend (e.g. "Asia").
	Continent string
	// Admin1 filters by a first-level administrative division. Discover valid
	// values with Client.GeoAdmin1.
	Admin1 string
	// Bbox is a sub-country viewport "lat_min,lon_min,lat_max,lon_max".
	Bbox string
	// Category filters by event category/domain, sent comma-separated.
	Category []string
	// Subcategory filters by sub-event type; requires Category.
	Subcategory string
	// StartDate / EndDate bound the window (ISO date); max 30 days.
	StartDate string
	EndDate   string
	// HasFatalities narrows to events with fatalities when set.
	HasFatalities    bool
	HasHasFatalities bool
	// CivilianTargeting narrows to conflict events targeting civilians when set.
	CivilianTargeting    bool
	HasCivilianTargeting bool
}

func (p EventsSummaryParams) values() url.Values {
	v := url.Values{}
	setStr(v, "group_by", p.GroupBy)
	setCSV(v, "country", p.Country)
	setStr(v, "region", p.Region)
	setStr(v, "continent", p.Continent)
	setStr(v, "admin1", p.Admin1)
	setStr(v, "bbox", p.Bbox)
	setCSV(v, "category", p.Category)
	setStr(v, "subcategory", p.Subcategory)
	setStr(v, "date_start", p.StartDate)
	setStr(v, "date_end", p.EndDate)
	setBool(v, "has_fatalities", p.HasFatalities, p.HasHasFatalities)
	setBool(v, "civilian_targeting", p.CivilianTargeting, p.HasCivilianTargeting)
	return v
}

// EventsSummary fetches grouped aggregate statistics for events
// (GET /api/v2/events/summary).
func (c *Client) EventsSummary(ctx context.Context, params EventsSummaryParams) ([]SummaryBucket, error) {
	var out []SummaryBucket
	if err := c.get(ctx, "/api/v2/events/summary", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EventsSummaryRaw fetches grouped event statistics and returns the complete
// response body verbatim, preserving the success envelope, the top-level
// group_by field and the full per-bucket statistics.
func (c *Client) EventsSummaryRaw(ctx context.Context, params EventsSummaryParams) (json.RawMessage, error) {
	return c.rawBody(ctx, "/api/v2/events/summary", params.values())
}
