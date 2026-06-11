# gdeltcloud-sdk

A Go SDK and command-line client for the [GDELT Cloud v2 API](https://docs.gdeltcloud.com/api-reference/v2).

It provides a small, typed client over the public GDELT Cloud endpoints:

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/v2/events` | `Client.Events` | Generated events |
| `/api/v2/stories` | `Client.Stories` | Story clusters |
| `/api/v2/entities` | `Client.Entities` | Entities (people, organizations, places) |
| `/api/v2/energy/assets` | `Client.EnergyAssets` | GEM-tracked energy assets in a bounding box |

Authentication uses a GDELT Cloud API key (format `gdelt_sk_...`) sent in the `Authorization` header using the `Bearer` scheme
(i.e. `Authorization: Bearer gdelt_sk_...`). Get a key at
<https://gdeltcloud.com/api-keys>.

## Installation

```sh
go get github.com/lonegunmanb/gdeltcloud-sdk
```

Requires Go 1.24+.

## Library usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	gdeltcloud "github.com/lonegunmanb/gdeltcloud-sdk"
)

func main() {
	client, err := gdeltcloud.NewClient("gdelt_sk_...")
	if err != nil {
		log.Fatal(err)
	}

	events, err := client.Events(context.Background(), gdeltcloud.EventsParams{
		Country:   []string{"YEM", "SAU"},
		StartDate: "2026-04-21",
		EndDate:   "2026-05-21",
		Limit:     50,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range events {
		fmt.Printf("%s — %s\n", e.EventDate, e.Title)
	}
}
```

### Configuration options

`NewClient` accepts functional options:

- `WithBaseURL(url)` — override the API base URL (default `https://gdeltcloud.com`).
- `WithHTTPClient(*http.Client)` — supply a custom HTTP client.
- `WithTimeout(d)` — set the request timeout on the default client (default 60s).
- `WithUserAgent(ua)` — override the `User-Agent` header.

### Error handling

API errors (an error envelope or a non-2xx status) are returned as
`*gdeltcloud.APIError`, which exposes `StatusCode`, `Message` and `Body`:

```go
events, err := client.Events(ctx, params)
var apiErr *gdeltcloud.APIError
if errors.As(err, &apiErr) {
	log.Printf("status %d: %s", apiErr.StatusCode, apiErr.Message)
}
```

> **Note:** the events endpoint caps the date window at 30 days per call. For
> longer windows, issue multiple calls and merge the results.

## Command-line client

The module ships a `gdelt` CLI under `cmd/gdelt`.

```sh
# install
go install github.com/lonegunmanb/gdeltcloud-sdk/cmd/gdelt@latest

# or build locally
go build -o gdelt ./cmd/gdelt
```

Provide your API key via the `GDELT_API_KEY` environment variable or the
`--api-key` flag.

```sh
export GDELT_API_KEY=gdelt_sk_...

gdelt events --country YEM,SAU --start 2026-04-21 --end 2026-05-21 --limit 50
gdelt stories --country YEM --start 2026-05-01 --end 2026-05-07 --article-count-min 4
gdelt entities --search Houthi --start 2026-05-01 --end 2026-05-07 --include-images
gdelt energy-assets --bbox 11.5,42.5,13.5,44.5 --tracker oil_gas_plants,lng_terminals
```

Output is JSON (indented by default; use `--compact` for single-line). Run
`gdelt help` for the full command list and `gdelt help <command>` (or
`gdelt <command> -h`) for per-command flags.

## Testing

Unit tests run against an in-process mock server and require no credentials:

```sh
go test ./...
```

An end-to-end test (`TestE2E`) exercises the live API. It is **skipped** unless
`GDELT_API_KEY` is set:

```sh
GDELT_API_KEY=gdelt_sk_... go test -run TestE2E -v ./...
```

## Continuous integration

The [`CI` workflow](.github/workflows/ci.yml) runs `go vet`, build, and unit
tests (with `-race`) on every push and pull request. A separate job runs the
live e2e test using the `GDELT_API_KEY` repository secret; if the secret is not
configured, the e2e test self-skips.

## License

See [LICENSE](LICENSE).
