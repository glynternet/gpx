package geojson

import (
	stdjson "encoding/json"
	"fmt"

	"github.com/glynternet/gpx/pkg/json"
)

// FeatureCollection is the subset of the GeoJSON (RFC 7946) FeatureCollection
// structure that is relevant to producing waypoints. Unknown fields are
// ignored so that the rich properties produced by upstream tooling (e.g. raw
// OSM tags) do not need to be modelled here.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string            `json:"type"`
	Geometry   Geometry          `json:"geometry"`
	Properties FeatureProperties `json:"properties"`
}

type Geometry struct {
	Type string `json:"type"`
	// Coordinates are in GeoJSON order: [longitude, latitude(, elevation)].
	Coordinates []float64 `json:"coordinates"`
}

type FeatureProperties struct {
	Name string `json:"name"`
	// Category is a single human-readable category, mapped to the waypoint
	// symbol as it is the closest equivalent field.
	Category   string   `json:"category"`
	Categories []string `json:"categories"`
	OSMID      int64    `json:"osmid"`
	// Tags carries the raw upstream (e.g. OSM) tags, serialised into the
	// waypoint description.
	Tags map[string]string `json:"tags"`
}

// Points converts the GeoJSON features into the flat waypoint representation
// shared with the custom JSON format. Only Point geometries are supported.
func (fc FeatureCollection) Points() ([]json.Point, error) {
	points := make([]json.Point, 0, len(fc.Features))
	for i, f := range fc.Features {
		if f.Geometry.Type != "Point" {
			return nil, fmt.Errorf("feature %d (%q): unsupported geometry type %q, only Point is supported", i, f.Properties.Name, f.Geometry.Type)
		}
		// RFC 7946: a Point position is [lon, lat] or [lon, lat, elevation].
		if l := len(f.Geometry.Coordinates); l != 2 && l != 3 {
			return nil, fmt.Errorf("feature %d (%q): expected 2 or 3 coordinates [lon, lat(, ele)], got %d", i, f.Properties.Name, l)
		}
		var desc string
		if len(f.Properties.Tags) > 0 {
			tags, err := stdjson.MarshalIndent(f.Properties.Tags, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("feature %d (%q): marshaling tags: %w", i, f.Properties.Name, err)
			}
			desc = string(tags)
		}
		points = append(points, json.Point{
			Name:        f.Properties.Name,
			Lon:         f.Geometry.Coordinates[0],
			Lat:         f.Geometry.Coordinates[1],
			Description: desc,
			Symbol:      f.Properties.Category,
			OSMID:       f.Properties.OSMID,
			Categories:  f.Properties.Categories,
		})
	}
	return points, nil
}
