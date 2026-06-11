// Package gdeltcloud is a Go SDK for the GDELT Cloud v2 REST API.
//
// It provides a small, typed client over the public endpoints documented at
// https://docs.gdeltcloud.com/api-reference/v2: events, stories, entities and
// energy assets (list endpoints), the matching fetch-by-id and summary
// endpoints, and the geo/admin1 division lookup. Authentication uses a GDELT
// Cloud API key (format
// "gdelt_sk_...") sent as an HTTP Bearer token in the Authorization header.
//
// Basic usage:
//
//	client, err := gdeltcloud.NewClient("gdelt_sk_...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	events, err := client.Events(context.Background(), gdeltcloud.EventsParams{
//	    Country:   []string{"YEM", "SAU"},
//	    StartDate: "2026-04-21",
//	    EndDate:   "2026-05-21",
//	    Limit:     50,
//	})
package gdeltcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production base URL for the GDELT Cloud API.
const DefaultBaseURL = "https://gdeltcloud.com"

// DefaultTimeout is the default per-request timeout used when no custom HTTP
// client is supplied.
const DefaultTimeout = 60 * time.Second

// Client is a GDELT Cloud API client. It is safe for concurrent use by
// multiple goroutines.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client. Options are applied in order by NewClient.
type Option func(*Client)

// WithBaseURL overrides the API base URL (default DefaultBaseURL). This is
// useful for testing against a mock server or a local development instance.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient sets a custom *http.Client. When provided, it takes precedence
// over WithTimeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithTimeout sets the timeout on the default HTTP client. It has no effect if
// a custom client is supplied via WithHTTPClient.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Timeout = d
		}
	}
}

// WithUserAgent overrides the User-Agent header sent on every request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// NewClient creates a new GDELT Cloud API client. The apiKey is required; get
// one at https://gdeltcloud.com/api-keys.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gdeltcloud: API key is required (get one at https://gdeltcloud.com/api-keys)")
	}
	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		userAgent:  "gdeltcloud-go-sdk",
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: DefaultTimeout}
	}
	return c, nil
}

// APIError represents an error returned by the GDELT Cloud API, either as an
// error envelope ({"success": false, "error": "..."}) or as a non-2xx HTTP
// status.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Message is the error message reported by the API, when available.
	Message string
	// Body is the raw response body, useful for debugging unexpected errors.
	Body string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("gdeltcloud: API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("gdeltcloud: API error (status %d): %s", e.StatusCode, e.Body)
}

// envelope is the standard GDELT Cloud response wrapper.
type envelope struct {
	Success *bool           `json:"success"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

// do performs a GET request against path with the given query values, validates
// the response envelope (HTTP status and the "success" flag) and returns the
// raw response body together with the parsed envelope.
//
// It is the shared transport used by both get (which extracts the "data" field)
// and endpoints such as /api/v2/geo/admin1 whose payload lives at the top level
// of the envelope rather than inside "data".
func (c *Client) do(ctx context.Context, path string, query url.Values) ([]byte, *envelope, error) {
	endpoint := c.baseURL + path
	if enc := query.Encode(); enc != "" {
		endpoint += "?" + enc
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("gdeltcloud: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("gdeltcloud: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("gdeltcloud: read response: %w", err)
	}

	var env envelope
	// The body should always be a JSON envelope, but guard against non-JSON
	// error pages (e.g. gateway errors) so we surface a useful message.
	if jsonErr := json.Unmarshal(body, &env); jsonErr != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, nil, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
		}
		return nil, nil, fmt.Errorf("gdeltcloud: decode response: %w", jsonErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, &APIError{StatusCode: resp.StatusCode, Message: env.Error, Body: string(body)}
	}
	if env.Success != nil && !*env.Success {
		return nil, nil, &APIError{StatusCode: resp.StatusCode, Message: env.Error, Body: string(body)}
	}

	return body, &env, nil
}

// get performs a GET request against path with the given query values and
// decodes the "data" field of the response envelope into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	_, env, err := c.do(ctx, path, query)
	if err != nil {
		return err
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("gdeltcloud: decode data: %w", err)
	}
	return nil
}
