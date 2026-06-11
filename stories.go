package gdeltcloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// StoriesParams are the query parameters for the stories endpoint.
type StoriesParams struct {
	// Country filters by country (ISO-3 codes), sent comma-separated.
	Country []string
	// StartDate is the inclusive start of the window (ISO date).
	StartDate string
	// EndDate is the inclusive end of the window (ISO date).
	EndDate string
	// ArticleCountMin filters out story clusters with fewer than this many
	// articles.
	ArticleCountMin int
	// Limit caps the number of returned records.
	Limit int
	// IncludeImages toggles article sharing-image enrichment.
	IncludeImages    bool
	HasIncludeImages bool
}

func (p StoriesParams) values() url.Values {
	v := url.Values{}
	setCSV(v, "country", p.Country)
	setStr(v, "date_start", p.StartDate)
	setStr(v, "date_end", p.EndDate)
	setInt(v, "article_count_min", p.ArticleCountMin)
	setInt(v, "limit", p.Limit)
	setBool(v, "include_images", p.IncludeImages, p.HasIncludeImages)
	return v
}

// Story is a single story cluster returned by the stories endpoint.
type Story struct {
	ID          string        `json:"id,omitempty"`
	ClusterID   string        `json:"cluster_id,omitempty"`
	Label       string        `json:"label,omitempty"`
	Title       string        `json:"title,omitempty"`
	URL         string        `json:"url,omitempty"`
	ClusterDate string        `json:"cluster_date,omitempty"`
	StoryDate   string        `json:"story_date,omitempty"`
	Category    string        `json:"category,omitempty"`
	Subcategory string        `json:"subcategory,omitempty"`
	ImageURL    string        `json:"image_url,omitempty"`
	Geo         *Geo          `json:"geo,omitempty"`
	Metrics     *StoryMetrics `json:"metrics,omitempty"`
	TopArticles []Article     `json:"top_articles,omitempty"`
}

// Stories fetches story clusters matching the given parameters.
func (c *Client) Stories(ctx context.Context, params StoriesParams) ([]Story, error) {
	var out []Story
	if err := c.get(ctx, "/api/v2/stories", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Story fetches a single story cluster by its v2 identifier
// (GET /api/v2/stories/{story_id}).
func (c *Client) Story(ctx context.Context, id string) (*Story, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("gdeltcloud: story id is required")
	}
	var out Story
	if err := c.get(ctx, "/api/v2/stories/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StoriesSummaryParams are the query parameters for the stories/summary
// endpoint. See EventsSummaryParams for the GroupBy semantics; the summary
// endpoint does not accept the free-text "search" parameter.
type StoriesSummaryParams struct {
	// GroupBy is the aggregation dimension (required): "date", "country",
	// "region", "continent", "category" or "subcategory".
	GroupBy string
	// Country filters by country, sent comma-separated.
	Country []string
	// Region expands to a country list on the backend.
	Region string
	// Continent expands to a country list on the backend.
	Continent string
	// Admin1 filters by a first-level administrative division.
	Admin1 string
	// Bbox is a sub-country viewport "lat_min,lon_min,lat_max,lon_max".
	Bbox string
	// Category filters by linked-event category, sent comma-separated.
	Category []string
	// Subcategory filters by linked sub-event type; requires Category.
	Subcategory string
	// StartDate / EndDate bound the window (ISO date); max 30 days.
	StartDate string
	EndDate   string
	// ArticleCountMin / ArticleCountMax bound the article evidence volume.
	ArticleCountMin int
	ArticleCountMax int
	// HasEvents narrows to stories with at least one linked event when set.
	HasEvents    bool
	HasHasEvents bool
	// HasFatalities narrows to stories with linked fatal events when set.
	HasFatalities    bool
	HasHasFatalities bool
}

func (p StoriesSummaryParams) values() url.Values {
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
	setInt(v, "article_count_min", p.ArticleCountMin)
	setInt(v, "article_count_max", p.ArticleCountMax)
	setBool(v, "has_events", p.HasEvents, p.HasHasEvents)
	setBool(v, "has_fatalities", p.HasFatalities, p.HasHasFatalities)
	return v
}

// StoriesSummary fetches grouped aggregate statistics for story clusters
// (GET /api/v2/stories/summary).
func (c *Client) StoriesSummary(ctx context.Context, params StoriesSummaryParams) ([]SummaryBucket, error) {
	var out []SummaryBucket
	if err := c.get(ctx, "/api/v2/stories/summary", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
