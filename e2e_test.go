package gdeltcloud_test

import (
	"context"
	"os"
	"testing"
	"time"

	gdeltcloud "github.com/lonegunmanb/gdeltcloud-sdk"
)

// TestE2E exercises the live GDELT Cloud API. It is skipped unless
// GDELT_API_KEY is set in the environment, so it is safe to run in CI without
// a secret and locally without credentials.
//
// Run it explicitly with:
//
//	GDELT_API_KEY=gdelt_sk_... go test -run TestE2E -v ./...
func TestE2E(t *testing.T) {
	apiKey := os.Getenv("GDELT_API_KEY")
	if apiKey == "" {
		t.Skip("GDELT_API_KEY not set; skipping live end-to-end test")
	}

	opts := []gdeltcloud.Option{}
	if base := os.Getenv("GDELT_BASE_URL"); base != "" {
		opts = append(opts, gdeltcloud.WithBaseURL(base))
	}
	client, err := gdeltcloud.NewClient(apiKey, opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Use a recent ~7 day window similar to the official demos.
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -7)
	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")
	countries := []string{"YEM", "DJI", "ERI", "SAU", "EGY", "ISR"}

	t.Run("events", func(t *testing.T) {
		events, err := client.Events(ctx, gdeltcloud.EventsParams{
			Country:   countries,
			StartDate: startDate,
			EndDate:   endDate,
			Limit:     10,
		})
		if err != nil {
			t.Fatalf("Events: %v", err)
		}
		t.Logf("events returned: %d", len(events))
	})

	t.Run("stories", func(t *testing.T) {
		stories, err := client.Stories(ctx, gdeltcloud.StoriesParams{
			Country:         countries,
			StartDate:       startDate,
			EndDate:         endDate,
			ArticleCountMin: 4,
			Limit:           10,
		})
		if err != nil {
			t.Fatalf("Stories: %v", err)
		}
		t.Logf("stories returned: %d", len(stories))
	})

	t.Run("entities", func(t *testing.T) {
		entities, err := client.Entities(ctx, gdeltcloud.EntitiesParams{
			Search:    "Houthi",
			StartDate: startDate,
			EndDate:   endDate,
			Limit:     10,
		})
		if err != nil {
			t.Fatalf("Entities: %v", err)
		}
		t.Logf("entities returned: %d", len(entities))
	})

	t.Run("energy_assets", func(t *testing.T) {
		assets, err := client.EnergyAssets(ctx, gdeltcloud.EnergyAssetsParams{
			Bbox:    "11.5,42.5,13.5,44.5",
			Tracker: []string{"oil_gas_plants", "lng_terminals"},
			Limit:   10,
		})
		if err != nil {
			t.Fatalf("EnergyAssets: %v", err)
		}
		t.Logf("energy assets returned: %d", len(assets))
	})
}
