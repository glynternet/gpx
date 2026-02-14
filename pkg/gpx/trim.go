package gpx

import (
	"fmt"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// DistanceIndex returns the index of the point at which the cumulative distance from the start has either been met or
// breached. In practise, that means that as long as there is no error, the index and index-1 both exist and are safe
// to use.
//
//		e.g.
//		For a slice of 9 points (index 0-8), pts, where pts[7] is at 100m and pts[8] is at 102m,
//		7 would be returned for a given distance of 100.
//		8 would be returned for a given distance of greater than 100, up to and including 102.
//	 	An error would be returned for a given distance of greater than 102.
func DistanceIndex(pts []gpxgo.GPXPoint, distance float64) (int, error) {
	var cumulativeDistance float64
	for i, pt := range pts {
		if i+1 >= len(pts) {
			return -1, fmt.Errorf("distance not reached for given points, total distance: %fm", cumulativeDistance)
		}
		cumulativeDistance += pt.Distance3D(&pts[i+1])
		if cumulativeDistance >= distance {
			return i + 1, nil
		}
	}
	return -1, fmt.Errorf("distance not reached for given points, total distance: %fm", cumulativeDistance)
}
