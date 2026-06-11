package gdeltcloud

// Geo describes a geographic location attached to events, stories and assets.
type Geo struct {
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Country   string   `json:"country,omitempty"`
	Name      string   `json:"name,omitempty"`
}

// EventMetrics holds the analytic scores attached to an event.
type EventMetrics struct {
	Magnitude            *float64 `json:"magnitude,omitempty"`
	GoldsteinScale       *float64 `json:"goldstein_scale,omitempty"`
	PropagationPotential *float64 `json:"propagation_potential,omitempty"`
	MarketSensitivity    *float64 `json:"market_sensitivity,omitempty"`
	SystemicImportance   *float64 `json:"systemic_importance,omitempty"`
	Fatalities           *float64 `json:"fatalities,omitempty"`
	Significance         *float64 `json:"significance,omitempty"`
}

// StoryMetrics holds counts attached to a story cluster.
type StoryMetrics struct {
	ArticleCount     *int `json:"article_count,omitempty"`
	LinkedEventCount *int `json:"linked_event_count,omitempty"`
	StoryCount       *int `json:"story_count,omitempty"`
}

// EntityMetrics holds counts attached to an entity.
type EntityMetrics struct {
	ArticleCount *int `json:"article_count,omitempty"`
	StoryCount   *int `json:"story_count,omitempty"`
	EventCount   *int `json:"event_count,omitempty"`
}

// Capacity describes the rated capacity of an energy asset.
type Capacity struct {
	MW *float64 `json:"mw,omitempty"`
}

// Wikipedia holds the Wikipedia enrichment returned for an entity when
// include_images is enabled.
type Wikipedia struct {
	Description  string `json:"description,omitempty"`
	PageURL      string `json:"page_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// Article is a single article referenced by a story cluster.
type Article struct {
	Title  string `json:"title,omitempty"`
	URL    string `json:"url,omitempty"`
	Domain string `json:"domain,omitempty"`
}
