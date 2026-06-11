// Command gdelt is a command-line client for the GDELT Cloud v2 API.
//
// It wraps the github.com/lonegunmanb/gdeltcloud-sdk package so the API can be
// called directly from the shell. Run "gdelt help" for usage.
package main

import (
	"fmt"
	"os"

	gdeltcloud "github.com/lonegunmanb/gdeltcloud-sdk"
)

// version is overridable at build time with -ldflags "-X main.version=...".
var version = "dev"

const rootUsage = `gdelt - command-line client for the GDELT Cloud v2 API

USAGE:
    gdelt <command> [flags]

COMMANDS:
    events            Query generated events
    event             Fetch a single event by id
    events-summary    Aggregate events into grouped summary buckets
    stories           Query story clusters
    story             Fetch a single story cluster by id
    stories-summary   Aggregate story clusters into grouped summary buckets
    entities          Query entities (people, organizations, places)
    entity            Fetch a single entity by id
    energy-assets     Query GEM-tracked energy assets within a bounding box
    admin1            List a country's first-level administrative divisions
    version           Print the gdelt version
    help              Show this help, or help for a command

AUTHENTICATION:
    A GDELT Cloud API key (format "gdelt_sk_...") is required. Provide it with
    the --api-key flag or the GDELT_API_KEY environment variable. Get a key at
    https://gdeltcloud.com/api-keys

GLOBAL FLAGS (available on every command):
    --api-key string     GDELT Cloud API key (env: GDELT_API_KEY)
    --base-url string    API base URL (env: GDELT_BASE_URL) (default %q)
    --timeout duration   HTTP request timeout (default %s)
    --compact            Emit compact single-line JSON instead of indented JSON

EXAMPLES:
    gdelt events --country YEM,SAU --start 2026-04-21 --end 2026-05-21 --limit 50
    gdelt event --id conflict_20260417_example
    gdelt events-summary --group-by country --region "Middle East" --start 2026-04-01 --end 2026-04-17
    gdelt stories --country YEM --start 2026-05-01 --end 2026-05-07 --article-count-min 4
    gdelt story --id story_20260417_example
    gdelt stories-summary --group-by date --continent Asia --start 2026-04-01 --end 2026-04-17
    gdelt entities --search Houthi --start 2026-05-01 --end 2026-05-07 --include-images
    gdelt entity --id person:Example%%20Person
    gdelt energy-assets --bbox 11.5,42.5,13.5,44.5 --tracker oil_gas_plants,lng_terminals
    gdelt admin1 --country France

Run "gdelt help <command>" for detailed flags of a specific command.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printRootUsage(os.Stderr)
		return 2
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "help", "-h", "--help":
		return cmdHelp(rest)
	case "version", "--version":
		fmt.Println("gdelt " + version)
		return 0
	case "events":
		return cmdEvents(rest)
	case "event":
		return cmdEvent(rest)
	case "events-summary":
		return cmdEventsSummary(rest)
	case "stories":
		return cmdStories(rest)
	case "story":
		return cmdStory(rest)
	case "stories-summary":
		return cmdStoriesSummary(rest)
	case "entities":
		return cmdEntities(rest)
	case "entity":
		return cmdEntity(rest)
	case "energy-assets":
		return cmdEnergyAssets(rest)
	case "admin1":
		return cmdAdmin1(rest)
	default:
		fmt.Fprintf(os.Stderr, "gdelt: unknown command %q\n\n", cmd)
		printRootUsage(os.Stderr)
		return 2
	}
}

func printRootUsage(w *os.File) {
	fmt.Fprintf(w, rootUsage, gdeltcloud.DefaultBaseURL, gdeltcloud.DefaultTimeout)
}

func cmdHelp(args []string) int {
	if len(args) == 0 {
		printRootUsage(os.Stdout)
		return 0
	}
	// Trigger the subcommand's own usage by passing -h.
	switch args[0] {
	case "events":
		return cmdEvents([]string{"-h"})
	case "event":
		return cmdEvent([]string{"-h"})
	case "events-summary":
		return cmdEventsSummary([]string{"-h"})
	case "stories":
		return cmdStories([]string{"-h"})
	case "story":
		return cmdStory([]string{"-h"})
	case "stories-summary":
		return cmdStoriesSummary([]string{"-h"})
	case "entities":
		return cmdEntities([]string{"-h"})
	case "entity":
		return cmdEntity([]string{"-h"})
	case "energy-assets":
		return cmdEnergyAssets([]string{"-h"})
	case "admin1":
		return cmdAdmin1([]string{"-h"})
	default:
		fmt.Fprintf(os.Stderr, "gdelt: unknown command %q\n", args[0])
		return 2
	}
}
