package elevation

import "sort"

// SplitRequest describes how to split a profile into segments.
type SplitRequest struct {
	Track     Profile   `json:"track"`
	Mode      string    `json:"mode"`      // "equidistant" or "waypoints"
	Count     int       `json:"count"`     // for equidistant mode
	Distances []float64 `json:"distances"` // for waypoints mode
}

// SplitResult contains the split segments and their distance boundaries.
type SplitResult struct {
	Segments   []Profile    `json:"segments"`
	Boundaries [][2]float64 `json:"boundaries"`
}

// SplitByDistance splits a profile into n equidistant segments.
func SplitByDistance(profile Profile, n int) SplitResult {
	if n <= 1 {
		totalDist := lastTrackpointDistance(profile.Track)
		return SplitResult{
			Segments:   []Profile{profile},
			Boundaries: [][2]float64{{0, totalDist}},
		}
	}

	totalDist := lastTrackpointDistance(profile.Track)
	segLen := totalDist / float64(n)

	segments := make([]Profile, n)
	boundaries := make([][2]float64, n)
	for i := 0; i < n; i++ {
		segStart := float64(i) * segLen
		segEnd := float64(i+1) * segLen
		segments[i] = buildSegment(profile, segStart, segEnd)
		boundaries[i] = [2]float64{segStart, segEnd}
	}
	return SplitResult{Segments: segments, Boundaries: boundaries}
}

// SplitByWaypoints splits a profile at the given distances.
func SplitByWaypoints(profile Profile, distances []float64) SplitResult {
	totalDist := lastTrackpointDistance(profile.Track)

	sorted := make([]float64, len(distances))
	copy(sorted, distances)
	sort.Float64s(sorted)

	bounds := make([]float64, 0, len(sorted)+2)
	bounds = append(bounds, 0)
	bounds = append(bounds, sorted...)
	bounds = append(bounds, totalDist)

	n := len(bounds) - 1
	segments := make([]Profile, n)
	boundaries := make([][2]float64, n)
	for i := 0; i < n; i++ {
		segments[i] = buildSegment(profile, bounds[i], bounds[i+1])
		boundaries[i] = [2]float64{bounds[i], bounds[i+1]}
	}
	return SplitResult{Segments: segments, Boundaries: boundaries}
}

func buildSegment(profile Profile, segStart, segEnd float64) Profile {
	segTrackpoints := extractSegmentTrackpoints(segStart, segEnd, profile.Track)

	var segWaypoints []Waypoint
	for _, w := range profile.Waypoints {
		if w.Distance >= segStart && w.Distance <= segEnd {
			wp := w
			wp.Distance -= segStart
			segWaypoints = append(segWaypoints, wp)
		}
	}

	// Normalize trackpoint distances relative to segment start.
	normalized := make([]TrackPoint, len(segTrackpoints))
	for i, tp := range segTrackpoints {
		normalized[i] = tp
		normalized[i].Distance -= segStart
	}

	var gain, loss float64
	if len(segTrackpoints) >= 2 {
		first := segTrackpoints[0]
		last := segTrackpoints[len(segTrackpoints)-1]
		gain = last.Gain - first.Gain
		loss = last.Loss - first.Loss
	}

	return Profile{
		Track:     normalized,
		Waypoints: segWaypoints,
		Gain:      gain,
		Loss:      loss,
	}
}

func extractSegmentTrackpoints(segStart, segEnd float64, trackpoints []TrackPoint) []TrackPoint {
	var pointsInRange []TrackPoint
	for _, tp := range trackpoints {
		if tp.Distance >= segStart && tp.Distance <= segEnd {
			pointsInRange = append(pointsInRange, tp)
		}
	}

	startPoint := interpolateTrackpointAt(segStart, trackpoints)
	endPoint := interpolateTrackpointAt(segEnd, trackpoints)

	// Prepend interpolated start if needed.
	withStart := pointsInRange
	if startPoint != nil {
		if len(pointsInRange) == 0 {
			withStart = []TrackPoint{*startPoint}
		} else if pointsInRange[0].Distance > segStart {
			withStart = append([]TrackPoint{*startPoint}, pointsInRange...)
		}
	}

	// Append interpolated end if needed.
	result := withStart
	if endPoint != nil {
		if len(withStart) == 0 {
			result = []TrackPoint{*endPoint}
		} else if withStart[len(withStart)-1].Distance < segEnd {
			result = append(withStart, *endPoint)
		}
	}

	return result
}

func interpolateTrackpointAt(dist float64, trackpoints []TrackPoint) *TrackPoint {
	if len(trackpoints) == 0 {
		return nil
	}
	if len(trackpoints) == 1 {
		if trackpoints[0].Distance == dist {
			tp := trackpoints[0]
			return &tp
		}
		return nil
	}

	for i := 0; i < len(trackpoints)-1; i++ {
		a := trackpoints[i]
		b := trackpoints[i+1]

		if a.Distance == dist {
			tp := a
			return &tp
		}
		if a.Distance < dist && b.Distance >= dist {
			var t float64
			if b.Distance != a.Distance {
				t = (dist - a.Distance) / (b.Distance - a.Distance)
			}
			return &TrackPoint{
				Distance:  dist,
				Elevation: a.Elevation + t*(b.Elevation-a.Elevation),
				Lat:       a.Lat + t*(b.Lat-a.Lat),
				Lon:       a.Lon + t*(b.Lon-a.Lon),
				Gain:      a.Gain + t*(b.Gain-a.Gain),
				Loss:      a.Loss + t*(b.Loss-a.Loss),
			}
		}
	}

	// Check if dist matches the last point exactly.
	last := trackpoints[len(trackpoints)-1]
	if last.Distance == dist {
		return &last
	}
	return nil
}

func lastTrackpointDistance(trackpoints []TrackPoint) float64 {
	if len(trackpoints) == 0 {
		return 0
	}
	return trackpoints[len(trackpoints)-1].Distance
}
