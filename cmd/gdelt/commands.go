package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	gdeltcloud "github.com/lonegunmanb/gdeltcloud-sdk"
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
func emit(cf *commonFlags, v any) int {
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
	includeImages := fs.Bool("include-images", false, "include image enrichment in results")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query generated events from the GDELT Cloud API.\n\n"+
			"Note: the events window is capped at 30 days per call.\n\n"+
			"USAGE:\n    gdelt events [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt events --country YEM,SAU --start 2026-04-21 --end 2026-05-21 --limit 50\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.Events(ctx, gdeltcloud.EventsParams{
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			Domain:           *domain,
			Limit:            *limit,
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
	includeImages := fs.Bool("include-images", false, "include article sharing-image enrichment")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query story clusters from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt stories [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt stories --country YEM --start 2026-05-01 --end 2026-05-07 --article-count-min 4\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.Stories(ctx, gdeltcloud.StoriesParams{
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			ArticleCountMin:  *articleCountMin,
			Limit:            *limit,
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
	includeImages := fs.Bool("include-images", false, "include Wikipedia thumbnail enrichment")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Query entities (people, organizations, places) from the GDELT Cloud API.\n\n"+
			"USAGE:\n    gdelt entities [flags]\n\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLE:\n    gdelt entities --search Houthi --start 2026-05-01 --end 2026-05-07 --include-images\n")
	}

	return runWithClient(fs, cf, args, func(ctx context.Context, c *gdeltcloud.Client) (any, error) {
		return c.Entities(ctx, gdeltcloud.EntitiesParams{
			Search:           *search,
			Country:          splitCSV(*country),
			StartDate:        *start,
			EndDate:          *end,
			Limit:            *limit,
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
		return c.EnergyAssets(ctx, gdeltcloud.EnergyAssetsParams{
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
