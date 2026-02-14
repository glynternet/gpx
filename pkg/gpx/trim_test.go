package gpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func TestDistanceIndex(t *testing.T) {
	t.Run("errors on no points", func(t *testing.T) {
		_, err := DistanceIndex(nil, 0)
		assert.EqualError(t, err, "distance not reached for given points, total distance: 0.000000m")
	})

	t.Run("errors on one point", func(t *testing.T) {
		_, err := DistanceIndex([]gpxgo.GPXPoint{{}}, 0)
		assert.EqualError(t, err, "distance not reached for given points, total distance: 0.000000m")
	})

	t.Run("errors when points don't reach distance", func(t *testing.T) {
		pt1 := gpxgo.GPXPoint{}
		pt2 := gpxgo.GPXPoint{
			Point: gpxgo.Point{Latitude: 1},
		}
		dist := pt1.Distance3D(&pt2)
		require.NotZero(t, dist)
		_, err := DistanceIndex([]gpxgo.GPXPoint{pt1, pt1, pt2}, dist+1)
		assert.EqualError(t, err, "distance not reached for given points, total distance: 111194.926645m")
	})

	t.Run("returns index when able to reach distance exactly", func(t *testing.T) {
		pt1 := gpxgo.GPXPoint{}
		pt2 := gpxgo.GPXPoint{
			Point: gpxgo.Point{Latitude: 1},
		}
		dist := pt1.Distance3D(&pt2)
		require.NotZero(t, dist)
		index, err := DistanceIndex([]gpxgo.GPXPoint{pt1, pt1, pt2}, dist)
		assert.NoError(t, err, "distance not reached for given points, total distance: 111194.926645m")
		assert.Equal(t, 2, index)
	})

	t.Run("returns index when able to reach distance by exceeding it", func(t *testing.T) {
		pt1 := gpxgo.GPXPoint{}
		pt2 := gpxgo.GPXPoint{
			Point: gpxgo.Point{Latitude: 1},
		}
		dist := pt1.Distance3D(&pt2)
		require.NotZero(t, dist)
		threshold := 0.1
		require.Greater(t, dist, threshold)
		index, err := DistanceIndex([]gpxgo.GPXPoint{pt1, pt1, pt2}, 0.1)
		assert.NoError(t, err, "distance not reached for given points, total distance: 111194.926645m")
		assert.Equal(t, 2, index)
	})
}
