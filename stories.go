package gdeltcloud

import (
	"context"
	"net/url"
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
	setStr(v, "start_date", p.StartDate)
	setStr(v, "end_date", p.EndDate)
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
