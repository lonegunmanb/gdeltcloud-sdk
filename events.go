package gdeltcloud

import (
	"context"
	"net/url"
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
	// IncludeImages toggles image enrichment. Use the SetIncludeImages helper
	// or set IncludeImages together with HasIncludeImages.
	IncludeImages    bool
	HasIncludeImages bool
}

func (p EventsParams) values() url.Values {
	v := url.Values{}
	setCSV(v, "country", p.Country)
	setStr(v, "start_date", p.StartDate)
	setStr(v, "end_date", p.EndDate)
	setStr(v, "domain", p.Domain)
	setInt(v, "limit", p.Limit)
	setBool(v, "include_images", p.IncludeImages, p.HasIncludeImages)
	return v
}

// Event is a single event returned by the events endpoint.
type Event struct {
	ID          string        `json:"id,omitempty"`
	EventUID    string        `json:"event_uid,omitempty"`
	Title       string        `json:"title,omitempty"`
	URL         string        `json:"url,omitempty"`
	EventDate   string        `json:"event_date,omitempty"`
	Date        string        `json:"date,omitempty"`
	Category    string        `json:"category,omitempty"`
	Subcategory string        `json:"subcategory,omitempty"`
	Family      string        `json:"family,omitempty"`
	EventFamily string        `json:"event_family,omitempty"`
	Domain      string        `json:"domain,omitempty"`
	ImageURL    string        `json:"image_url,omitempty"`
	Geo         *Geo          `json:"geo,omitempty"`
	Metrics     *EventMetrics `json:"metrics,omitempty"`
}

// Events fetches events matching the given parameters.
func (c *Client) Events(ctx context.Context, params EventsParams) ([]Event, error) {
	var out []Event
	if err := c.get(ctx, "/api/v2/events", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
