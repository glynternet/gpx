package geojson

import (
	"testing"

	"github.com/glynternet/gpx/pkg/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureCollectionPoints(t *testing.T) {
	t.Run("maps fields and coordinate order", func(t *testing.T) {
		fc := FeatureCollection{
			Type: "FeatureCollection",
			Features: []Feature{{
				Type: "Feature",
				Geometry: Geometry{
					Type:        "Point",
					Coordinates: []float64{-106.286971, 36.493906},
				},
				Properties: FeatureProperties{
					Name:       "Banco Lona",
					Category:   "Summit",
					Categories: []string{"Summit"},
					OSMID:      357591944,
				},
			}},
		}

		ps, err := fc.Points()
		require.NoError(t, err)
		assert.Equal(t, []json.Point{{
			Name:       "Banco Lona",
			Lat:        36.493906,
			Lon:        -106.286971,
			Symbol:     "Summit",
			OSMID:      357591944,
			Categories: []string{"Summit"},
		}}, ps)
	})

	t.Run("serialises tags into the description", func(t *testing.T) {
		fc := FeatureCollection{Features: []Feature{{
			Geometry:   Geometry{Type: "Point", Coordinates: []float64{1, 2}},
			Properties: FeatureProperties{Tags: map[string]string{"natural": "peak"}},
		}}}
		ps, err := fc.Points()
		require.NoError(t, err)
		assert.Equal(t, "{\n  \"natural\": \"peak\"\n}", ps[0].Description)
	})

	t.Run("leaves description empty when there are no tags", func(t *testing.T) {
		fc := FeatureCollection{Features: []Feature{{
			Geometry: Geometry{Type: "Point", Coordinates: []float64{1, 2}},
		}}}
		ps, err := fc.Points()
		require.NoError(t, err)
		assert.Empty(t, ps[0].Description)
	})

	t.Run("accepts elevation as third coordinate", func(t *testing.T) {
		fc := FeatureCollection{Features: []Feature{{
			Geometry: Geometry{Type: "Point", Coordinates: []float64{1, 2, 3}},
		}}}
		ps, err := fc.Points()
		require.NoError(t, err)
		assert.Equal(t, 1.0, ps[0].Lon)
		assert.Equal(t, 2.0, ps[0].Lat)
	})

	t.Run("empty collection yields empty points", func(t *testing.T) {
		ps, err := FeatureCollection{}.Points()
		require.NoError(t, err)
		assert.Empty(t, ps)
	})

	t.Run("rejects non-Point geometry", func(t *testing.T) {
		fc := FeatureCollection{Features: []Feature{{
			Geometry: Geometry{Type: "LineString", Coordinates: []float64{1, 2}},
		}}}
		_, err := fc.Points()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "LineString")
	})

	t.Run("rejects coordinate lengths other than 2 or 3", func(t *testing.T) {
		for _, coords := range [][]float64{{}, {1}, {1, 2, 3, 4}} {
			fc := FeatureCollection{Features: []Feature{{
				Geometry: Geometry{Type: "Point", Coordinates: coords},
			}}}
			_, err := fc.Points()
			require.Error(t, err, "coordinates: %v", coords)
		}
	})
}
