package gdeltcloud

import (
	"context"
	"net/url"
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
	// IncludeImages toggles Wikipedia thumbnail enrichment.
	IncludeImages    bool
	HasIncludeImages bool
}

func (p EntitiesParams) values() url.Values {
	v := url.Values{}
	setStr(v, "search", p.Search)
	setCSV(v, "country", p.Country)
	setStr(v, "start_date", p.StartDate)
	setStr(v, "end_date", p.EndDate)
	setInt(v, "limit", p.Limit)
	setBool(v, "include_images", p.IncludeImages, p.HasIncludeImages)
	return v
}

// Entity is a single entity returned by the entities endpoint.
type Entity struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Label        string         `json:"label,omitempty"`
	URL          string         `json:"url,omitempty"`
	WikipediaURL string         `json:"wikipedia_url,omitempty"`
	ImageURL     string         `json:"image_url,omitempty"`
	AvatarURL    string         `json:"avatar_url,omitempty"`
	ThumbnailURL string         `json:"thumbnail_url,omitempty"`
	Wikipedia    *Wikipedia     `json:"wikipedia,omitempty"`
	Metrics      *EntityMetrics `json:"metrics,omitempty"`
}

// Entities fetches entities matching the given parameters.
func (c *Client) Entities(ctx context.Context, params EntitiesParams) ([]Entity, error) {
	var out []Entity
	if err := c.get(ctx, "/api/v2/entities", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
