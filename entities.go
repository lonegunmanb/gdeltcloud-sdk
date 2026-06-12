package gdeltcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// EntitiesParams are the query parameters for the entities endpoint.
type EntitiesParams struct {
	// Search is the entity search/anchor term (e.g. "Houthi").
	Search string
	// Country filters by country (ISO-3 codes), sent comma-separated.
	Country []string
	// StartDate is the inclusive start of the window (ISO date).
	StartDate string
	// EndDate is the inclusive end of the window (ISO date).
	EndDate string
	// Limit caps the number of returned records.
	Limit int
	// Cursor is the pagination cursor: pass the next_cursor value from a prior
	// response's pagination block to fetch the next page.
	Cursor string
	// IncludeImages toggles Wikipedia thumbnail enrichment.
	IncludeImages    bool
	HasIncludeImages bool
}

func (p EntitiesParams) values() url.Values {
	v := url.Values{}
	setStr(v, "search", p.Search)
	setCSV(v, "country", p.Country)
	setStr(v, "date_start", p.StartDate)
	setStr(v, "date_end", p.EndDate)
	setInt(v, "limit", p.Limit)
	setStr(v, "cursor", p.Cursor)
	setBool(v, "include_images", p.IncludeImages, p.HasIncludeImages)
	return v
}

// Entity is a single entity returned by the entities endpoint. It models the
// documented v2 Entity card and detail record (StoryRefs / EventRefs are
// populated only by the detail endpoint); callers that need byte-for-byte
// fidelity to the API response can use EntitiesRaw / EntityRaw.
type Entity struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Label        string         `json:"label,omitempty"`
	Type         string         `json:"type,omitempty"`
	URL          string         `json:"url,omitempty"`
	WikipediaURL string         `json:"wikipedia_url,omitempty"`
	ImageURL     string         `json:"image_url,omitempty"`
	AvatarURL    string         `json:"avatar_url,omitempty"`
	ThumbnailURL string         `json:"thumbnail_url,omitempty"`
	LatestDate   string         `json:"latest_date,omitempty"`
	Wikipedia    *Wikipedia     `json:"wikipedia,omitempty"`
	Metrics      *EntityMetrics `json:"metrics,omitempty"`
	StoryRefs    []StoryRef     `json:"story_refs,omitempty"`
	EventRefs    []EventRef     `json:"event_refs,omitempty"`
}

// Entities fetches entities matching the given parameters.
func (c *Client) Entities(ctx context.Context, params EntitiesParams) ([]Entity, error) {
	var out []Entity
	if err := c.get(ctx, "/api/v2/entities", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EntitiesRaw fetches entities and returns the complete response body verbatim,
// preserving the full success envelope (success, data and pagination) and every
// documented record field.
func (c *Client) EntitiesRaw(ctx context.Context, params EntitiesParams) (json.RawMessage, error) {
	return c.rawBody(ctx, "/api/v2/entities", params.values())
}

// Entity fetches a single entity by its v2 identifier
// (GET /api/v2/entities/{entity_id}).
func (c *Client) Entity(ctx context.Context, id string) (*Entity, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gdeltcloud: entity id is required")
	}
	var out Entity
	if err := c.get(ctx, "/api/v2/entities/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EntityRaw fetches a single entity by id and returns the complete response
// body verbatim. See EntitiesRaw for the rationale.
func (c *Client) EntityRaw(ctx context.Context, id string) (json.RawMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gdeltcloud: entity id is required")
	}
	return c.rawBody(ctx, "/api/v2/entities/"+url.PathEscape(id), nil)
}
