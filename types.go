package gdeltcloud

import "encoding/json"

// Geo describes a geographic location attached to events, stories and assets.
type Geo struct {
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Country   string   `json:"country,omitempty"`
	Region    string   `json:"region,omitempty"`
	Continent string   `json:"continent,omitempty"`
	Admin1    string   `json:"admin1,omitempty"`
	Location  string   `json:"location,omitempty"`
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
	Summary      string `json:"summary,omitempty"`
	PageURL      string `json:"page_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// Article is a single article referenced by a story cluster.
type Article struct {
	Title  string `json:"title,omitempty"`
	URL    string `json:"url,omitempty"`
	Domain string `json:"domain,omitempty"`
}

// SummaryBucket is a single grouped bucket returned by the events/summary and
// stories/summary endpoints. The populated fields depend on the family being
// summarized (events vs. stories) and on the selected group_by dimension; the
// nested Metrics and MetricStats objects are preserved as raw JSON so callers
// always receive the full statistic set the API returns.
type SummaryBucket struct {
	// Key is the bucket value for the selected group_by (e.g. a country name
	// or a date). Country keys are normalized to plain English names.
	Key string `json:"key,omitempty"`
	// GroupBy is the concrete grouping dimension echoed by the API.
	GroupBy string `json:"group_by,omitempty"`

	// Counts shared by event and story summaries.
	EventCount   *int `json:"event_count,omitempty"`
	StoryCount   *int `json:"story_count,omitempty"`
	CountryCount *int `json:"country_count,omitempty"`
	RegionCount  *int `json:"region_count,omitempty"`

	// Event-summary specific counts.
	ConflictEventCount  *int     `json:"conflict_event_count,omitempty"`
	CameoplusEventCount *int     `json:"cameoplus_event_count,omitempty"`
	FatalityEventCount  *int     `json:"fatality_event_count,omitempty"`
	Fatalities          *int     `json:"fatalities,omitempty"`
	FatalityEventRate   *float64 `json:"fatality_event_rate,omitempty"`

	// Article-evidence aggregates.
	ArticleCount    *int     `json:"article_count,omitempty"`
	AvgArticleCount *float64 `json:"avg_article_count,omitempty"`
	MinArticleCount *int     `json:"min_article_count,omitempty"`
	MaxArticleCount *int     `json:"max_article_count,omitempty"`

	// Flat significance aggregates retained for simple clients.
	AvgSignificance   *float64 `json:"avg_significance,omitempty"`
	MaxSignificance   *float64 `json:"max_significance,omitempty"`
	MinSignificance   *float64 `json:"min_significance,omitempty"`
	AvgGoldsteinScale *float64 `json:"avg_goldstein_scale,omitempty"`

	// Story-summary specific fields.
	StoriesWithFatalities *int     `json:"stories_with_fatalities,omitempty"`
	FatalityStoryRate     *float64 `json:"fatality_story_rate,omitempty"`
	StoriesWithEvents     *int     `json:"stories_with_events,omitempty"`
	StoryOnlyCount        *int     `json:"story_only_count,omitempty"`
	LinkedEventCount      *int     `json:"linked_event_count,omitempty"`
	MinLinkedEventCount   *int     `json:"min_linked_event_count,omitempty"`
	AvgLinkedEventCount   *float64 `json:"avg_linked_event_count,omitempty"`
	MaxLinkedEventCount   *int     `json:"max_linked_event_count,omitempty"`
	AvgRecencyScore       *float64 `json:"avg_recency_score,omitempty"`

	// Metrics holds the nested aggregate statistics for significance and its
	// input metrics. MetricStats is an alias the API returns for clients that
	// prefer explicit statistical naming. Both are preserved verbatim.
	Metrics     json.RawMessage `json:"metrics,omitempty"`
	MetricStats json.RawMessage `json:"metric_stats,omitempty"`
}

// Admin1 is the response of the /api/v2/geo/admin1 endpoint: the list of valid
// first-level administrative divisions (states/provinces) for a country.
type Admin1 struct {
	// Country is the resolved country name the divisions belong to.
	Country string `json:"country,omitempty"`
	// Admin1 is the list of first-level administrative division names.
	Admin1 []string `json:"admin1,omitempty"`
	// Source identifies the dataset the divisions were derived from.
	Source string `json:"source,omitempty"`
}
