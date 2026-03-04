package elevation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpolateTrackpointAt(t *testing.T) {
	trackpoints := []TrackPoint{
		{Distance: 0, Elevation: 100, Lat: 50, Lon: -1, Gain: 0, Loss: 0},
		{Distance: 100, Elevation: 200, Lat: 51, Lon: 0, Gain: 100, Loss: 0},
		{Distance: 200, Elevation: 150, Lat: 52, Lon: 1, Gain: 100, Loss: 50},
	}

	t.Run("empty list", func(t *testing.T) {
		assert.Nil(t, interpolateTrackpointAt(50, nil))
	})

	t.Run("single point exact match", func(t *testing.T) {
		tp := interpolateTrackpointAt(0, trackpoints[:1])
		require.NotNil(t, tp)
		assert.Equal(t, 0.0, tp.Distance)
		assert.Equal(t, 100.0, tp.Elevation)
	})

	t.Run("single point no match", func(t *testing.T) {
		assert.Nil(t, interpolateTrackpointAt(50, trackpoints[:1]))
	})

	t.Run("exact match first point", func(t *testing.T) {
		tp := interpolateTrackpointAt(0, trackpoints)
		require.NotNil(t, tp)
		assert.Equal(t, 0.0, tp.Distance)
	})

	t.Run("exact match last point", func(t *testing.T) {
		tp := interpolateTrackpointAt(200, trackpoints)
		require.NotNil(t, tp)
		assert.Equal(t, 200.0, tp.Distance)
	})

	t.Run("midpoint interpolation", func(t *testing.T) {
		tp := interpolateTrackpointAt(50, trackpoints)
		require.NotNil(t, tp)
		assert.Equal(t, 50.0, tp.Distance)
		assert.Equal(t, 150.0, tp.Elevation)
		assert.Equal(t, 50.5, tp.Lat)
		assert.Equal(t, -0.5, tp.Lon)
		assert.Equal(t, 50.0, tp.Gain)
		assert.Equal(t, 0.0, tp.Loss)
	})

	t.Run("beyond range", func(t *testing.T) {
		assert.Nil(t, interpolateTrackpointAt(300, trackpoints))
	})

	t.Run("before range", func(t *testing.T) {
		assert.Nil(t, interpolateTrackpointAt(-10, trackpoints))
	})
}

func TestExtractSegmentTrackpoints(t *testing.T) {
	trackpoints := []TrackPoint{
		{Distance: 0, Elevation: 100},
		{Distance: 100, Elevation: 200},
		{Distance: 200, Elevation: 150},
		{Distance: 300, Elevation: 250},
	}

	t.Run("full range", func(t *testing.T) {
		result := extractSegmentTrackpoints(0, 300, trackpoints)
		assert.Len(t, result, 4)
	})

	t.Run("sub-range with interpolation", func(t *testing.T) {
		result := extractSegmentTrackpoints(50, 250, trackpoints)
		require.Len(t, result, 4)
		assert.Equal(t, 50.0, result[0].Distance)
		assert.Equal(t, 150.0, result[0].Elevation)
		assert.Equal(t, 250.0, result[3].Distance)
		assert.Equal(t, 200.0, result[3].Elevation)
	})

	t.Run("exact boundaries match existing points", func(t *testing.T) {
		result := extractSegmentTrackpoints(100, 200, trackpoints)
		assert.Len(t, result, 2)
		assert.Equal(t, 100.0, result[0].Distance)
		assert.Equal(t, 200.0, result[1].Distance)
	})
}

func testProfile() Profile {
	return Profile{
		Track: []TrackPoint{
			{Distance: 0, Elevation: 100, Lat: 50, Lon: -1, Gain: 0, Loss: 0},
			{Distance: 100, Elevation: 200, Lat: 51, Lon: 0, Gain: 100, Loss: 0},
			{Distance: 200, Elevation: 150, Lat: 52, Lon: 1, Gain: 100, Loss: 50},
			{Distance: 300, Elevation: 250, Lat: 53, Lon: 2, Gain: 200, Loss: 50},
		},
		Waypoints: []Waypoint{
			{Name: "Start", Distance: 0, Gain: 0, Loss: 0},
			{Name: "Mid", Distance: 150, Gain: 100, Loss: 0},
			{Name: "End", Distance: 300, Gain: 200, Loss: 50},
		},
		Gain: 200,
		Loss: 50,
	}
}

func TestSplitByDistance(t *testing.T) {
	profile := testProfile()

	t.Run("n=1 returns original", func(t *testing.T) {
		result := SplitByDistance(profile, 1)
		require.Len(t, result.Segments, 1)
		assert.Equal(t, profile, result.Segments[0])
		assert.Equal(t, [2]float64{0, 300}, result.Boundaries[0])
	})

	t.Run("n=3 splits into 3 equal segments", func(t *testing.T) {
		result := SplitByDistance(profile, 3)
		require.Len(t, result.Segments, 3)
		require.Len(t, result.Boundaries, 3)

		assert.Equal(t, [2]float64{0, 100}, result.Boundaries[0])
		assert.Equal(t, [2]float64{100, 200}, result.Boundaries[1])
		assert.Equal(t, [2]float64{200, 300}, result.Boundaries[2])

		// First segment trackpoints should start at distance 0
		assert.Equal(t, 0.0, result.Segments[0].Track[0].Distance)

		// Each segment's trackpoints are normalized to start from 0
		for _, seg := range result.Segments {
			assert.Equal(t, 0.0, seg.Track[0].Distance)
		}
	})

	t.Run("segment waypoints are filtered and normalized", func(t *testing.T) {
		result := SplitByDistance(profile, 3)
		// "Start" (dist=0) is in segment 0
		assert.Len(t, result.Segments[0].Waypoints, 1)
		assert.Equal(t, "Start", result.Segments[0].Waypoints[0].Name)
		assert.Equal(t, 0.0, result.Segments[0].Waypoints[0].Distance)

		// "Mid" (dist=150) is in segment 1 (100-200), normalized to 50
		assert.Len(t, result.Segments[1].Waypoints, 1)
		assert.Equal(t, "Mid", result.Segments[1].Waypoints[0].Name)
		assert.Equal(t, 50.0, result.Segments[1].Waypoints[0].Distance)

		// "End" (dist=300) is in segment 2
		assert.Len(t, result.Segments[2].Waypoints, 1)
		assert.Equal(t, "End", result.Segments[2].Waypoints[0].Name)
		assert.Equal(t, 100.0, result.Segments[2].Waypoints[0].Distance)
	})

	t.Run("gain/loss computed per segment", func(t *testing.T) {
		result := SplitByDistance(profile, 3)
		// Segment 0: dist 0-100, gain 0->100, loss 0->0
		assert.Equal(t, 100.0, result.Segments[0].Gain)
		assert.Equal(t, 0.0, result.Segments[0].Loss)
		// Segment 1: dist 100-200, gain 100->100, loss 0->50
		assert.Equal(t, 0.0, result.Segments[1].Gain)
		assert.Equal(t, 50.0, result.Segments[1].Loss)
		// Segment 2: dist 200-300, gain 100->200, loss 50->50
		assert.Equal(t, 100.0, result.Segments[2].Gain)
		assert.Equal(t, 0.0, result.Segments[2].Loss)
	})
}

func TestSplitByWaypoints(t *testing.T) {
	profile := testProfile()

	t.Run("split at distance 150", func(t *testing.T) {
		result := SplitByWaypoints(profile, []float64{150})
		require.Len(t, result.Segments, 2)
		assert.Equal(t, [2]float64{0, 150}, result.Boundaries[0])
		assert.Equal(t, [2]float64{150, 300}, result.Boundaries[1])

		// First segment normalized distances start at 0
		assert.Equal(t, 0.0, result.Segments[0].Track[0].Distance)
		assert.Equal(t, 0.0, result.Segments[1].Track[0].Distance)
	})

	t.Run("unsorted distances are sorted", func(t *testing.T) {
		result := SplitByWaypoints(profile, []float64{200, 100})
		require.Len(t, result.Segments, 3)
		assert.Equal(t, [2]float64{0, 100}, result.Boundaries[0])
		assert.Equal(t, [2]float64{100, 200}, result.Boundaries[1])
		assert.Equal(t, [2]float64{200, 300}, result.Boundaries[2])
	})

	t.Run("empty distances returns single segment", func(t *testing.T) {
		result := SplitByWaypoints(profile, nil)
		require.Len(t, result.Segments, 1)
		assert.Equal(t, [2]float64{0, 300}, result.Boundaries[0])
	})
}

func TestBuildSegment(t *testing.T) {
	profile := testProfile()

	t.Run("full range preserves all data", func(t *testing.T) {
		seg := buildSegment(profile, 0, 300)
		assert.Len(t, seg.Track, 4)
		assert.Len(t, seg.Waypoints, 3)
		assert.Equal(t, 200.0, seg.Gain)
		assert.Equal(t, 50.0, seg.Loss)
	})

	t.Run("sub-range interpolates boundaries", func(t *testing.T) {
		seg := buildSegment(profile, 50, 250)
		// Should have: interpolated@50, 100, 200, interpolated@250
		require.Len(t, seg.Track, 4)
		assert.Equal(t, 0.0, seg.Track[0].Distance)   // 50-50=0
		assert.Equal(t, 50.0, seg.Track[1].Distance)   // 100-50=50
		assert.Equal(t, 150.0, seg.Track[2].Distance)  // 200-50=150
		assert.Equal(t, 200.0, seg.Track[3].Distance)  // 250-50=200
	})
}
