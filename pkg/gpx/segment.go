package gpx

import (
	"errors"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// SplitPoints splits a slice of points into a number of segments.
// Pre-overlap percentage is the percentage of the segment size that should be prepended to the next segment.
func SplitPoints(points []gpxgo.GPXPoint, segments uint, preoverlapPercentage float32) ([][]gpxgo.GPXPoint, error) {
	if segments == 0 {
		return nil, errors.New("segments must be greater than 0")
	}
	if preoverlapPercentage < 0 || preoverlapPercentage > 100 {
		return nil, errors.New("preoverlapPercentage must be between 0-100, inclusive")
	}
	out := make([][]gpxgo.GPXPoint, segments)
	segmentSize := len(points) / int(segments)
	if segmentSize < 1 {
		segmentSize = 1
	}
	preOverlapSize := int(preoverlapPercentage / 100 * float32(segmentSize))
	for i := 0; i < int(segments); i++ {
		segmentStart := i * segmentSize
		segmentEnd := segmentStart + segmentSize
		if segmentEnd > len(points) {
			segmentEnd = len(points)
		}
		preoverlapStart := segmentStart - preOverlapSize
		if preoverlapStart < 0 {
			preoverlapStart = 0
		}
		out[i] = points[preoverlapStart:segmentEnd]
	}
	return out, nil
}
