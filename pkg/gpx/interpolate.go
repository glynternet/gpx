package gpx

import (
	"fmt"
	"slices"
	"time"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// PointsAt returns a point for each of the given times, linearly interpolated between the recorded
// points either side of it. Results are ordered by time, regardless of the order the times were
// given in. Each returned point carries only a position and the time it was requested for.
//
// pts must be non-empty, must all carry a timestamp and must be in ascending time order.
// Every time must fall within the times of the first and last of pts.
func PointsAt(pts []gpxgo.GPXPoint, times []time.Time) ([]gpxgo.GPXPoint, error) {
	if err := validateAscendingTimestamps(pts); err != nil {
		return nil, err
	}
	if len(times) == 0 {
		return nil, nil
	}

	sorted := slices.Clone(times)
	slices.SortFunc(sorted, func(a, b time.Time) int { return a.Compare(b) })

	first := pts[0].Timestamp
	last := pts[len(pts)-1].Timestamp
	if t := sorted[0]; t.Before(first) {
		return nil, fmt.Errorf("time %s is before first recorded point time %s",
			format(t), format(first))
	}
	if t := sorted[len(sorted)-1]; t.After(last) {
		return nil, fmt.Errorf("time %s is after last recorded point time %s",
			format(t), format(last))
	}

	out := make([]gpxgo.GPXPoint, 0, len(sorted))
	// i is the index of the first point at or after the time being resolved. Because the times
	// ascend, it only ever advances. It stays in bounds because no time is after the last point.
	var i int
	for _, t := range sorted {
		for pts[i].Timestamp.Before(t) {
			i++
		}
		if pts[i].Timestamp.Equal(t) {
			out = append(out, gpxgo.GPXPoint{Point: pts[i].Point, Timestamp: t})
			continue
		}
		// i > 0 here: t is not before the first point and the times are not equal, so the first
		// point at or after t cannot be the first of pts.
		out = append(out, interpolate(&pts[i-1], &pts[i], t))
	}
	return out, nil
}

// interpolate returns the point at the given time, which must be between the times of a and b.
func interpolate(a, b *gpxgo.GPXPoint, t time.Time) gpxgo.GPXPoint {
	fraction := float64(t.Sub(a.Timestamp)) / float64(b.Timestamp.Sub(a.Timestamp))
	point := gpxgo.Point{
		Latitude:  between(a.Latitude, b.Latitude, fraction),
		Longitude: between(a.Longitude, b.Longitude, fraction),
	}
	// elevation is only known for the interpolated point when it is known for both of its neighbours
	if a.Elevation.NotNull() && b.Elevation.NotNull() {
		point.Elevation = *gpxgo.NewNullableFloat64(
			between(a.Elevation.Value(), b.Elevation.Value(), fraction))
	}
	return gpxgo.GPXPoint{Point: point, Timestamp: t}
}

func between(a, b, fraction float64) float64 {
	return a + fraction*(b-a)
}

func validateAscendingTimestamps(pts []gpxgo.GPXPoint) error {
	if len(pts) == 0 {
		return fmt.Errorf("no points given")
	}
	for i := range pts {
		if pts[i].Timestamp.IsZero() {
			return fmt.Errorf("point %d has no timestamp", i)
		}
		if i > 0 && pts[i].Timestamp.Before(pts[i-1].Timestamp) {
			return fmt.Errorf("point %d time %s is before point %d time %s",
				i, format(pts[i].Timestamp), i-1, format(pts[i-1].Timestamp))
		}
	}
	return nil
}

func format(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
