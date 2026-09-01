package tcx

import (
	"fmt"
	"time"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// LapBounds is the first and last recorded position of a single lap.
type LapBounds struct {
	// Number is the lap's 1-based position in document order across the whole
	// file. It is kept even when earlier laps are skipped, so that waypoint
	// names refer to the lap the reader would count in the source.
	Number      int
	First, Last gpxgo.GPXPoint
}

// LapBounds returns the bounds of every lap in the document that recorded at
// least one position, alongside the numbers of those that recorded none.
//
// Laps of all activities are concatenated in document order and numbered
// continuously. Multi-sport files need no special handling: separate
// recordings do not share a boundary position, so Waypoints keeps their laps
// apart of its own accord.
func (db TrainingCenterDatabase) LapBounds() (bounds []LapBounds, skipped []int, err error) {
	var number int
	for _, activity := range db.Activities {
		for _, lap := range activity.Laps {
			number++
			first, last, found := lap.positioned()
			if !found {
				skipped = append(skipped, number)
				continue
			}
			firstPoint, err := first.gpxPoint()
			if err != nil {
				return nil, nil, fmt.Errorf("first point of lap %d: %w", number, err)
			}
			lastPoint, err := last.gpxPoint()
			if err != nil {
				return nil, nil, fmt.Errorf("last point of lap %d: %w", number, err)
			}
			bounds = append(bounds, LapBounds{
				Number: number,
				First:  firstPoint,
				Last:   lastPoint,
			})
		}
	}
	return bounds, skipped, nil
}

// positioned returns the lap's first and last trackpoints that carry a
// position, searching across all of its tracks. found is false when the lap
// recorded no position at all.
func (l Lap) positioned() (first, last Trackpoint, found bool) {
	for _, track := range l.Tracks {
		for _, tp := range track.Trackpoints {
			if tp.Position == nil {
				continue
			}
			if !found {
				first = tp
				found = true
			}
			last = tp
		}
	}
	return first, last, found
}

// gpxPoint converts a positioned trackpoint to a GPX point. It must only be
// called on a trackpoint with a position.
func (t Trackpoint) gpxPoint() (gpxgo.GPXPoint, error) {
	timestamp, err := time.Parse(time.RFC3339, t.Time)
	if err != nil {
		return gpxgo.GPXPoint{}, fmt.Errorf("parsing time %q: %w", t.Time, err)
	}
	point := gpxgo.GPXPoint{
		Point: gpxgo.Point{
			Latitude:  t.Position.Latitude,
			Longitude: t.Position.Longitude,
		},
		Timestamp: timestamp,
	}
	if t.Altitude != nil {
		point.Elevation = *gpxgo.NewNullableFloat64(*t.Altitude)
	}
	return point, nil
}
