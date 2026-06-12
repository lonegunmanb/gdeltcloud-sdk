package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	gdeltcloud "github.com/lonegunmanb/gdeltcloud-sdk"
)

// Sentinel errors for missing required per-command flags.
var (
	errMissingID      = errors.New("an --id is required")
	errMissingGroupBy = errors.New("a --group-by dimension is required (date|country|region|continent|category|subcategory)")
	errMissingCountry = errors.New("a --country is required")
)

// commonFlags holds the global flags shared by every subcommand.
type commonFlags struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	compact bool
}

// registerCommon attaches the global flags to a FlagSet.
func registerCommon(fs *flag.FlagSet) *commonFlags {
	cf := &commonFlags{}
	fs.StringVar(&cf.apiKey, "api-key", os.Getenv("GDELT_API_KEY"), "GDELT Cloud API key (env: GDELT_API_KEY)")
	fs.StringVar(&cf.baseURL, "base-url", os.Getenv("GDELT_BASE_URL"), "API base URL (env: GDELT_BASE_URL)")
	fs.DurationVar(&cf.timeout, "timeout", gdeltcloud.DefaultTimeout, "HTTP request timeout")
	fs.BoolVar(&cf.compact, "compact", false, "emit compact single-line JSON instead of indented JSON")
	return cf
}

// newClient builds a client from the resolved common flags.
func (cf *commonFlags) newClient() (*gdeltcloud.Client, error) {
	opts := []gdeltcloud.Option{gdeltcloud.WithTimeout(cf.timeout)}
	if cf.baseURL != "" {
		opts = append(opts, gdeltcloud.WithBaseURL(cf.baseURL))
	}
	return gdeltcloud.NewClient(cf.apiKey, opts...)
}

// usageError prints msg and the FlagSet usage, then returns exit code 2.
func usageError(fs *flag.FlagSet, msg string) int {
	fmt.Fprintf(os.Stderr, "gdelt: %s\n\n", msg)
	fs.Usage()
	return 2
}

// emit writes v as JSON to stdout, honoring the --compact flag.
//
// When v is a json.RawMessage (the verbatim API response body returned by the
// SDK's *Raw methods), it is re-formatted in place so the full documented
// envelope — success, data, pagination, group_by, and every record field — is
// preserved byte-for-byte without round-tripping through a lossy struct.
func emit(cf *commonFlags, v any) int {
	if raw, ok := v.(json.RawMessage); ok {
		var buf bytes.Buffer
		var err error
		if cf.compact {
			err = json.Compact(&buf, raw)
		} else {
			err = json.Indent(&buf, raw, "", "  ")
		}
		if err != nil {
			// Fall back to emitting the raw bytes unchanged rather than failing.
			buf.Reset()
			buf.Write(raw)
		}
		buf.WriteByte('\n')
		if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "gdelt: writing output: %v\n", err)
			return 1
		}
		return 0
	}

	enc := json.NewEncoder(os.Stdout)
	if !cf.compact {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "gdelt: encoding output: %v\n", err)
		return 1
	}
	return 0
}

// runWithClient parses flags, builds the client and runs fn. It centralizes
// API-key validation and error reporting.
func runWithClient(fs *flag.FlagSet, cf *commonFlags, args []string, fn func(context.Context, *gdeltcloud.Client) (any, error)) int {
	if err := fs.Parse(args); err != nil {
		// flag already printed the error / usage for -h.
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if cf.apiKey == "" {
		return usageError(fs, "an API key is required: set --api-key or GDELT_API_KEY")
	}
	client, err := cf.newClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gdelt: %v\n", err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), cf.timeout)
	defer cancel()

	result, err := fn(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gdelt: %v\n", err)
		return 1
	}
	return emit(cf, result)
}

// splitCSV splits a comma-separated flag value into a trimmed, non-empty slice.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func cmdEvents(args []string) int {
	fs := flag.NewFlagSet("events", flag.ContinueOnError)
	cf := registerCommon(fs)
	country := fs.String("country", "", "comma-separated ISO-3 country codes (e.g. YEM,SAU)")
	start := fs.String("start", "", "inclusive start date, ISO format YYYY-MM-DD")
	end := fs.String("end", "", "inclusive end date, ISO format YYYY-MM-DD")
	domain := fs.String("domain", "", "filter by CAMEO+ domain (e.g. INFRASTRUCTURE)")
	limit := fs.Int("limit", 0, "maximum number of records to return")
	cursor := fs.String("cursor", "", "pagination cursor (pass next_cursor from a prior response)")
	includeImages := fs.Bool("include-images", false, "include image enrichment in results")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query generated events from the GDELT Cloud API.\n\n"+
			"Note: the events window is capped at 30 days per call.\n\n"+
			"USAGE:\n    gdelt events [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt events --country YEM,SAU --start 2026-04-21 --end 2026-05-21 --limit 50\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.EventsRaw(ctx, gdeltcloud.EventsParams{
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			Domain:           *domain,
			Limit:            *limit,
			Cursor:           *cursor,
			IncludeImages:    *includeImages,
			HasIncludeImages: isSet(fs, "include-images"),
		})
	})
}

func cmdStories(args []string) int {
	fs := flag.NewFlagSet("stories", flag.ContinueOnError)
	cf := registerCommon(fs)
	country := fs.String("country", "", "comma-separated ISO-3 country codes (e.g. YEM,SAU)")
	start := fs.String("start", "", "inclusive start date, ISO format YYYY-MM-DD")
	end := fs.String("end", "", "inclusive end date, ISO format YYYY-MM-DD")
	articleCountMin := fs.Int("article-count-min", 0, "minimum article count per story cluster")
	limit := fs.Int("limit", 0, "maximum number of records to return")
	cursor := fs.String("cursor", "", "pagination cursor (pass next_cursor from a prior response)")
	includeImages := fs.Bool("include-images", false, "include article sharing-image enrichment")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query story clusters from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt stories [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt stories --country YEM --start 2026-05-01 --end 2026-05-07 --article-count-min 4\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.StoriesRaw(ctx, gdeltcloud.StoriesParams{
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			ArticleCountMin:  *articleCountMin,
			Limit:            *limit,
			Cursor:           *cursor,
			IncludeImages:    *includeImages,
			HasIncludeImages: isSet(fs, "include-images"),
		})
	})
}

func cmdEntities(args []string) int {
	fs := flag.NewFlagSet("entities", flag.ContinueOnError)
	cf := registerCommon(fs)
	search := fs.String("search", "", "entity search/anchor term (e.g. Houthi)")
	country := fs.String("country", "", "comma-separated ISO-3 country codes (e.g. YEM,SAU)")
	start := fs.String("start", "", "inclusive start date, ISO format YYYY-MM-DD")
	end := fs.String("end", "", "inclusive end date, ISO format YYYY-MM-DD")
	limit := fs.Int("limit", 0, "maximum number of records to return")
	cursor := fs.String("cursor", "", "pagination cursor (pass next_cursor from a prior response)")
	includeImages := fs.Bool("include-images", false, "include Wikipedia thumbnail enrichment")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query entities (people, organizations, places) from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt entities [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt entities --search Houthi --start 2026-05-01 --end 2026-05-07 --include-images\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.EntitiesRaw(ctx, gdeltcloud.EntitiesParams{
			Search:           *search,
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			Limit:            *limit,
			Cursor:           *cursor,
			IncludeImages:    *includeImages,
			HasIncludeImages: isSet(fs, "include-images"),
		})
	})
}

func cmdEnergyAssets(args []string) int {
	fs := flag.NewFlagSet("energy-assets", flag.ContinueOnError)
	cf := registerCommon(fs)
	bbox := fs.String("bbox", "", "bounding box lat_min,lon_min,lat_max,lon_max")
	tracker := fs.String("tracker", "", "comma-separated GEM trackers (e.g. oil_gas_plants,lng_terminals)")
	limit := fs.Int("limit", 0, "maximum number of records to return")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query GEM-tracked energy assets within a bounding box.\n\n"+
			"USAGE:\n    gdelt energy-assets [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt energy-assets --bbox 11.5,42.5,13.5,44.5 --tracker oil_gas_plants,lng_terminals\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.EnergyAssetsRaw(ctx, gdeltcloud.EnergyAssetsParams{
			Bbox:    *bbox,
			Tracker: splitCSV(*tracker),
			Limit:   *limit,
		})
	})
}

// isSet reports whether the named flag was explicitly provided on the command
// line (as opposed to left at its default).
func isSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func cmdEvent(args []string) int {
	fs := flag.NewFlagSet("event", flag.ContinueOnError)
	cf := registerCommon(fs)
	id := fs.String("id", "", "event id to fetch (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch a single event by its id from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt event --id <event_id> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt event --id conflict_20260417_example\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*id) == "" {
			return nil, errMissingID
		}
		return c.EventRaw(ctx, *id)
	})
}

func cmdStory(args []string) int {
	fs := flag.NewFlagSet("story", flag.ContinueOnError)
	cf := registerCommon(fs)
	id := fs.String("id", "", "story id to fetch (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch a single story cluster by its id from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt story --id <story_id> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt story --id story_20260417_example\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*id) == "" {
			return nil, errMissingID
		}
		return c.StoryRaw(ctx, *id)
	})
}

func cmdEntity(args []string) int {
	fs := flag.NewFlagSet("entity", flag.ContinueOnError)
	cf := registerCommon(fs)
	id := fs.String("id", "", "entity id to fetch (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch a single entity by its id from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt entity --id <entity_id> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt entity --id person:Example%%20Person\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*id) == "" {
			return nil, errMissingID
		}
		return c.EntityRaw(ctx, *id)
	})
}

func cmdEventsSummary(args []string) int {
	fs := flag.NewFlagSet("events-summary", flag.ContinueOnError)
	cf := registerCommon(fs)
	groupBy := fs.String("group-by", "", "aggregation dimension: date|country|region|continent|category|subcategory (required)")
	country := fs.String("country", "", "comma-separated country names/codes (e.g. France or YEM,SAU)")
	region := fs.String("region", "", "region name (e.g. Middle East)")
	continent := fs.String("continent", "", "continent name (e.g. Asia)")
	admin1 := fs.String("admin1", "", "first-level administrative division (discover via 'gdelt admin1')")
	bbox := fs.String("bbox", "", "bounding box lat_min,lon_min,lat_max,lon_max")
	category := fs.String("category", "", "comma-separated event categories (e.g. Protests,INFRASTRUCTURE)")
	subcategory := fs.String("subcategory", "", "sub-event type; requires --category")
	start := fs.String("start", "", "inclusive start date, ISO format YYYY-MM-DD")
	end := fs.String("end", "", "inclusive end date, ISO format YYYY-MM-DD")
	hasFatalities := fs.Bool("has-fatalities", false, "limit to events with fatalities")
	civilianTargeting := fs.Bool("civilian-targeting", false, "limit to conflict events targeting civilians")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Aggregate events into grouped summary buckets.\n\n"+
			"USAGE:\n    gdelt events-summary --group-by <dimension> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt events-summary --group-by country --region \"Middle East\" --has-fatalities --start 2026-04-01 --end 2026-04-17\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*groupBy) == "" {
			return nil, errMissingGroupBy
		}
		return c.EventsSummaryRaw(ctx, gdeltcloud.EventsSummaryParams{
			GroupBy:              *groupBy,
			Country:              splitCSV(*country),
			Region:               *region,
			Continent:            *continent,
			Admin1:               *admin1,
			Bbox:                 *bbox,
			Category:             splitCSV(*category),
			Subcategory:          *subcategory,
			StartDate:            *start,
			EndDate:              *end,
			HasFatalities:        *hasFatalities,
			HasHasFatalities:     isSet(fs, "has-fatalities"),
			CivilianTargeting:    *civilianTargeting,
			HasCivilianTargeting: isSet(fs, "civilian-targeting"),
		})
	})
}

func cmdStoriesSummary(args []string) int {
	fs := flag.NewFlagSet("stories-summary", flag.ContinueOnError)
	cf := registerCommon(fs)
	groupBy := fs.String("group-by", "", "aggregation dimension: date|country|region|continent|category|subcategory (required)")
	country := fs.String("country", "", "comma-separated country names/codes (e.g. France or YEM,SAU)")
	region := fs.String("region", "", "region name (e.g. Middle East)")
	continent := fs.String("continent", "", "continent name (e.g. Asia)")
	admin1 := fs.String("admin1", "", "first-level administrative division (discover via 'gdelt admin1')")
	bbox := fs.String("bbox", "", "bounding box lat_min,lon_min,lat_max,lon_max")
	category := fs.String("category", "", "comma-separated linked-event categories")
	subcategory := fs.String("subcategory", "", "linked sub-event type; requires --category")
	start := fs.String("start", "", "inclusive start date, ISO format YYYY-MM-DD")
	end := fs.String("end", "", "inclusive end date, ISO format YYYY-MM-DD")
	articleCountMin := fs.Int("article-count-min", 0, "minimum article count per story cluster")
	articleCountMax := fs.Int("article-count-max", 0, "maximum article count per story cluster")
	hasEvents := fs.Bool("has-events", false, "limit to stories with at least one linked event")
	hasFatalities := fs.Bool("has-fatalities", false, "limit to stories with linked fatal events")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Aggregate story clusters into grouped summary buckets.\n\n"+
			"USAGE:\n    gdelt stories-summary --group-by <dimension> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt stories-summary --group-by date --continent Asia --start 2026-04-01 --end 2026-04-17\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*groupBy) == "" {
			return nil, errMissingGroupBy
		}
		return c.StoriesSummaryRaw(ctx, gdeltcloud.StoriesSummaryParams{
			GroupBy:          *groupBy,
			Country:          splitCSV(*country),
			Region:           *region,
			Continent:        *continent,
			Admin1:           *admin1,
			Bbox:             *bbox,
			Category:         splitCSV(*category),
			Subcategory:      *subcategory,
			StartDate:        *start,
			EndDate:          *end,
			ArticleCountMin:  *articleCountMin,
			ArticleCountMax:  *articleCountMax,
			HasEvents:        *hasEvents,
			HasHasEvents:     isSet(fs, "has-events"),
			HasFatalities:    *hasFatalities,
			HasHasFatalities: isSet(fs, "has-fatalities"),
		})
	})
}

func cmdAdmin1(args []string) int {
	fs := flag.NewFlagSet("admin1", flag.ContinueOnError)
	cf := registerCommon(fs)
	country := fs.String("country", "", "country to list administrative divisions for (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "List the first-level administrative divisions (states/provinces) of a country.\n\n"+
			"USAGE:\n    gdelt admin1 --country <country> [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt admin1 --country France\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		if strings.TrimSpace(*country) == "" {
			return nil, errMissingCountry
		}
		return c.GeoAdmin1Raw(ctx, *country)
	})
}
