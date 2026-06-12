package gdeltcloud

import (
	"context"
	"encoding/json"
	"net/url"
)

// EnergyAssetsParams are the query parameters for the energy assets endpoint.
type EnergyAssetsParams struct {
	// Bbox is the bounding box "lat_min,lon_min,lat_max,lon_max".
	Bbox string
	// Tracker filters by GEM tracker(s) (e.g. "oil_gas_plants,lng_terminals"),
	// sent comma-separated.
	Tracker []string
	// Limit caps the number of returned records.
	Limit int
}

func (p EnergyAssetsParams) values() url.Values {
	v := url.Values{}
	setStr(v, "bbox", p.Bbox)
	setCSV(v, "tracker", p.Tracker)
	setInt(v, "limit", p.Limit)
	return v
}

// EnergyAsset is a single energy asset returned by the energy assets endpoint.
type EnergyAsset struct {
	ID       string    `json:"id,omitempty"`
	GemID    string    `json:"gem_id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Tracker  string    `json:"tracker,omitempty"`
	Geo      *Geo      `json:"geo,omitempty"`
	Capacity *Capacity `json:"capacity,omitempty"`
}

// EnergyAssets fetches energy assets matching the given parameters.
func (c *Client) EnergyAssets(ctx context.Context, params EnergyAssetsParams) ([]EnergyAsset, error) {
	var out []EnergyAsset
	if err := c.get(ctx, "/api/v2/energy/assets", params.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EnergyAssetsRaw fetches energy assets and returns the complete response body
// verbatim, preserving the full success envelope and every documented record
// field.
func (c *Client) EnergyAssetsRaw(ctx context.Context, params EnergyAssetsParams) (json.RawMessage, error) {
	return c.rawBody(ctx, "/api/v2/energy/assets", params.values())
}
