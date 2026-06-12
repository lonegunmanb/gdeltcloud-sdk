package gdeltcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GeoAdmin1 returns the valid first-level administrative divisions
// (states/provinces) for a country (GET /api/v2/geo/admin1?country=...).
//
// Use it to discover valid Admin1 values before filtering events or stories by
// admin1. The country argument accepts plain English names (e.g. "France"), as
// well as the ISO-3 and legacy FIPS aliases the backend supports.
//
// Unlike the list and detail endpoints, the admin1 payload lives at the top
// level of the response envelope rather than inside its "data" field, so the
// raw envelope is decoded directly into an Admin1.
func (c *Client) GeoAdmin1(ctx context.Context, country string) (*Admin1, error) {
	if strings.TrimSpace(country) == "" {
		return nil, fmt.Errorf("gdeltcloud: country is required")
	}
	q := url.Values{}
	q.Set("country", country)

	body, _, err := c.do(ctx, "/api/v2/geo/admin1", q)
	if err != nil {
		return nil, err
	}
	var out Admin1
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("gdeltcloud: decode data: %w", err)
	}
	return &out, nil
}

// GeoAdmin1Raw returns the complete /api/v2/geo/admin1 response body verbatim
// (the top-level {success, country, admin1, source} payload), preserving every
// field the API returns.
func (c *Client) GeoAdmin1Raw(ctx context.Context, country string) (json.RawMessage, error) {
	if strings.TrimSpace(country) == "" {
		return nil, fmt.Errorf("gdeltcloud: country is required")
	}
	q := url.Values{}
	q.Set("country", country)
	return c.rawBody(ctx, "/api/v2/geo/admin1", q)
}
