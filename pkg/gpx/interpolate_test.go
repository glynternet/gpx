package gpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func TestPointsAt(t *testing.T) {
	start := time.Date(2023, time.November, 14, 22, 13, 20, 0, time.UTC)
	at := func(seconds int) time.Time {
		return start.Add(time.Duration(seconds) * time.Second)
	}
	// a track of 3 points, 60s apart, moving 1 degree and 100m of elevation between each
	track := []gpxgo.GPXPoint{
		point(1, 10, 100, at(0)),
		point(2, 20, 200, at(60)),
		point(3, 30, 300, at(120)),
	}

	t.Run("errors on no points", func(t *testing.T) {
		_, err := PointsAt(nil, []time.Time{at(0)})
		assert.EqualError(t, err, "no points given")
	})

	t.Run("errors on point without timestamp", func(t *testing.T) {
		pts := []gpxgo.GPXPoint{track[0], {}}
		_, err := PointsAt(pts, []time.Time{at(0)})
		assert.EqualError(t, err, "point 1 has no timestamp")
	})

	t.Run("errors on points not in ascending time order", func(t *testing.T) {
		pts := []gpxgo.GPXPoint{track[1], track[0]}
		_, err := PointsAt(pts, []time.Time{at(0)})
		assert.EqualError(t, err,
			"point 1 time 2023-11-14T22:13:20Z is before point 0 time 2023-11-14T22:14:20Z")
	})

	t.Run("returns no points for no times", func(t *testing.T) {
		points, err := PointsAt(track, nil)
		assert.NoError(t, err)
		assert.Empty(t, points)
	})

	t.Run("errors on time before first recorded point", func(t *testing.T) {
		_, err := PointsAt(track, []time.Time{at(30), at(-1)})
		assert.EqualError(t, err,
			"time 2023-11-14T22:13:19Z is before first recorded point time 2023-11-14T22:13:20Z")
	})

	t.Run("errors on time after last recorded point", func(t *testing.T) {
		_, err := PointsAt(track, []time.Time{at(121), at(30)})
		assert.EqualError(t, err,
			"time 2023-11-14T22:15:21Z is after last recorded point time 2023-11-14T22:15:20Z")
	})

	t.Run("returns recorded point for exact time match", func(t *testing.T) {
		for i, pt := range track {
			points, err := PointsAt(track, []time.Time{pt.Timestamp})
			require.NoError(t, err, "point %d", i)
			require.Len(t, points, 1)
			assert.Equal(t, pt.Point, points[0].Point)
			assert.Equal(t, pt.Timestamp, points[0].Timestamp)
		}
	})

	t.Run("interpolates halfway between two points", func(t *testing.T) {
		points, err := PointsAt(track, []time.Time{at(30)})
		require.NoError(t, err)
		require.Len(t, points, 1)
		assert.Equal(t, point(1.5, 15, 150, at(30)), points[0])
	})

	t.Run("interpolates quarter of the way between two points", func(t *testing.T) {
		points, err := PointsAt(track, []time.Time{at(75)})
		require.NoError(t, err)
		require.Len(t, points, 1)
		assert.Equal(t, point(2.25, 22.5, 225, at(75)), points[0])
	})

	t.Run("elevation is null when a neighbouring point has none", func(t *testing.T) {
		pts := []gpxgo.GPXPoint{track[0], {
			Point:     gpxgo.Point{Latitude: 2, Longitude: 20},
			Timestamp: at(60),
		}}
		points, err := PointsAt(pts, []time.Time{at(30)})
		require.NoError(t, err)
		require.Len(t, points, 1)
		assert.True(t, points[0].Elevation.Null())
		assert.Equal(t, 1.5, points[0].Latitude)
		assert.Equal(t, 15.0, points[0].Longitude)
	})

	t.Run("returns points in time order for unsorted times", func(t *testing.T) {
		points, err := PointsAt(track, []time.Time{at(90), at(30), at(90)})
		require.NoError(t, err)
		assert.Equal(t, []gpxgo.GPXPoint{
			point(1.5, 15, 150, at(30)),
			point(2.5, 25, 250, at(90)),
			point(2.5, 25, 250, at(90)),
		}, points)
	})

	t.Run("resolves many times spanning the whole track in a single call", func(t *testing.T) {
		points, err := PointsAt(track, []time.Time{at(0), at(30), at(60), at(90), at(120)})
		require.NoError(t, err)
		assert.Equal(t, []gpxgo.GPXPoint{
			point(1, 10, 100, at(0)),
			point(1.5, 15, 150, at(30)),
			point(2, 20, 200, at(60)),
			point(2.5, 25, 250, at(90)),
			point(3, 30, 300, at(120)),
		}, points)
	})
}

func point(lat, lon, ele float64, timestamp time.Time) gpxgo.GPXPoint {
	return gpxgo.GPXPoint{
		Point: gpxgo.Point{
			Latitude:  lat,
			Longitude: lon,
			Elevation: *gpxgo.NewNullableFloat64(ele),
		},
		Timestamp: timestamp,
	}
}
