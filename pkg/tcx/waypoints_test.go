package tcx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func TestWaypoints(t *testing.T) {
	t.Run("merges a boundary the next lap begins at", func(t *testing.T) {
		ws := Waypoints([]LapBounds{
			{Number: 1, First: point(1, 1), Last: point(2, 2)},
			{Number: 2, First: point(2, 2), Last: point(3, 3)},
		})

		require.Len(t, ws, 3)
		assert.Equal(t, "Lap 1 (start)", ws[0].Name)
		assert.Equal(t, "Start of lap 1", ws[0].Description)
		assert.Equal(t, symbolStart, ws[0].Symbol)

		assert.Equal(t, "Lap 1 (finish) / Lap 2 (start)", ws[1].Name)
		assert.Equal(t, "End of lap 1, start of lap 2", ws[1].Description)
		assert.Equal(t, symbolBoundary, ws[1].Symbol)
		assert.Equal(t, 2.0, ws[1].Latitude)

		assert.Equal(t, "Lap 2 (finish)", ws[2].Name)
		assert.Equal(t, "End of lap 2", ws[2].Description)
		assert.Equal(t, symbolFinish, ws[2].Symbol)
	})

	t.Run("keeps apart a boundary the next lap does not begin at", func(t *testing.T) {
		// the recording was paused at the end of lap 1 and resumed elsewhere
		ws := Waypoints([]LapBounds{
			{Number: 1, First: point(1, 1), Last: point(2, 2)},
			{Number: 2, First: point(9, 9), Last: point(3, 3)},
		})

		require.Len(t, ws, 4)
		assert.Equal(t, "Lap 1 (start)", ws[0].Name)
		assert.Equal(t, "Lap 1 (finish)", ws[1].Name)
		assert.Equal(t, symbolFinish, ws[1].Symbol)
		assert.Equal(t, 2.0, ws[1].Latitude)
		assert.Equal(t, "Lap 2 (start)", ws[2].Name)
		assert.Equal(t, symbolStart, ws[2].Symbol)
		assert.Equal(t, 9.0, ws[2].Latitude)
		assert.Equal(t, "Lap 2 (finish)", ws[3].Name)
	})

	t.Run("does not merge positions that share only a latitude", func(t *testing.T) {
		ws := Waypoints([]LapBounds{
			{Number: 1, First: point(1, 1), Last: point(2, 2)},
			{Number: 2, First: point(2, 9), Last: point(3, 3)},
		})
		require.Len(t, ws, 4)
	})

	t.Run("names by lap number, not position, when a lap was skipped", func(t *testing.T) {
		ws := Waypoints([]LapBounds{
			{Number: 1, First: point(1, 1), Last: point(2, 2)},
			{Number: 3, First: point(2, 2), Last: point(3, 3)},
		})
		require.Len(t, ws, 3)
		assert.Equal(t, "Lap 1 (finish) / Lap 3 (start)", ws[1].Name)
		assert.Equal(t, "Lap 3 (finish)", ws[2].Name)
	})

	t.Run("gives a single lap a start and a finish", func(t *testing.T) {
		ws := Waypoints([]LapBounds{{Number: 1, First: point(1, 1), Last: point(2, 2)}})
		require.Len(t, ws, 2)
		assert.Equal(t, "Lap 1 (start)", ws[0].Name)
		assert.Equal(t, 1.0, ws[0].Latitude)
		assert.Equal(t, "Lap 1 (finish)", ws[1].Name)
		assert.Equal(t, 2.0, ws[1].Latitude)
	})

	t.Run("gives every waypoint the user type", func(t *testing.T) {
		ws := Waypoints([]LapBounds{
			{Number: 1, First: point(1, 1), Last: point(2, 2)},
			{Number: 2, First: point(9, 9), Last: point(3, 3)},
		})
		for _, w := range ws {
			assert.Equal(t, "user", w.Type, w.Name)
		}
	})

	t.Run("returns nothing for no laps", func(t *testing.T) {
		assert.Nil(t, Waypoints(nil))
	})
}

func point(lat, lon float64) gpxgo.GPXPoint {
	return gpxgo.GPXPoint{Point: gpxgo.Point{Latitude: lat, Longitude: lon}}
}
