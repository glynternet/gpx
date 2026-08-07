package elevation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkrajina/gpxgo/gpx"
)

func TestCategoriesFromSymbol(t *testing.T) {
	for _, test := range []struct {
		name   string
		symbol string
		want   []string
	}{
		{name: "no symbol", symbol: "", want: nil},
		{name: "meaningful symbol", symbol: "Water Source", want: []string{"Water Source"}},
		{name: "default route planner symbol", symbol: "Dot", want: nil},
		{name: "default garmin symbol", symbol: "Flag, Blue", want: nil},
		{name: "uninformative symbol in any case", symbol: "dot", want: nil},
		{name: "uninformative symbol padded", symbol: " Flag, Blue ", want: nil},
		{name: "symbol that merely contains an uninformative one", symbol: "Dot, Red", want: []string{"Dot, Red"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, categoriesFromSymbol(test.symbol))
		})
	}
}

func TestCalculateProfilesWaypointCategories(t *testing.T) {
	parsed, err := gpx.ParseBytes([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="test" xmlns="http://www.topografix.com/GPX/1/1">
  <wpt lat="50.0" lon="-1.0">
    <name>categorised</name>
    <sym>Dot</sym>
    <extensions><Categories><Category>Cafe</Category></Categories></extensions>
  </wpt>
  <wpt lat="50.0" lon="-1.0">
    <name>symbol only</name>
    <sym>Water Source</sym>
  </wpt>
  <wpt lat="50.0" lon="-1.0">
    <name>default symbol</name>
    <sym>Dot</sym>
  </wpt>
  <wpt lat="50.0" lon="-1.0">
    <name>no symbol</name>
  </wpt>
  <trk><trkseg>
    <trkpt lat="50.0" lon="-1.0"><ele>100</ele></trkpt>
    <trkpt lat="51.0" lon="-1.0"><ele>200</ele></trkpt>
  </trkseg></trk>
</gpx>`))
	require.NoError(t, err)

	profiles, err := CalculateProfiles(parsed.Tracks, parsed.Waypoints)
	require.NoError(t, err)
	require.Len(t, profiles, 1)

	categoriesByName := map[string][]string{}
	for _, waypoint := range profiles[0].Waypoints {
		categoriesByName[waypoint.Name] = waypoint.Categories
	}
	assert.Equal(t, map[string][]string{
		"categorised":    {"Cafe"},
		"symbol only":    {"Water Source"},
		"default symbol": nil,
		"no symbol":      nil,
	}, categoriesByName)
}
