package gpx

import (
	"errors"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// SplitPoints divides a route into the requested number of chunks, at most one
// per segment.
//
// It is the segments that are divided, not the points: each chunk ends on the
// point the next begins with, so every segment lies inside some chunk and no
// point is left out. Segment counts differ by at most one.
//
// preoverlapPercentage reaches each chunk further back, over that percentage of
// its own length, giving neighbours more common ground than the single shared
// point. The first chunk has nothing to reach back into.
func SplitPoints(points []gpxgo.GPXPoint, chunks uint, preoverlapPercentage float32) ([][]gpxgo.GPXPoint, error) {
	if chunks == 0 {
		return nil, errors.New("chunks must be greater than 0")
	}
	// Asserted positively so that NaN, which compares false against everything,
	// is rejected rather than let through to make nonsense of the boundaries.
	if !(preoverlapPercentage >= 0 && preoverlapPercentage <= 100) {
		return nil, errors.New("preoverlapPercentage must be between 0-100, inclusive")
	}
	if len(points) == 0 {
		return nil, nil
	}
	if len(points) == 1 {
		// No segment to divide, but the point is still a route.
		return [][]gpxgo.GPXPoint{points}, nil
	}

	segments := len(points) - 1
	// Compared in uint space so that no chunks value converts to a negative int.
	resolvedNumChunks := int(min(uint(segments), chunks))

	out := make([][]gpxgo.GPXPoint, 0, resolvedNumChunks)
	for i := range resolvedNumChunks {
		// Proportional boundaries divide the segments evenly and leave each chunk
		// starting on the point its predecessor ended on. The +1 takes the point
		// closing the chunk's last segment.
		start := i * segments / resolvedNumChunks
		end := (i + 1) * segments / resolvedNumChunks
		overlap := int(preoverlapPercentage / 100 * float32(end-start))
		out = append(out, points[max(0, start-overlap):end+1])
	}
	return out, nil
}
