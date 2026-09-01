package tcx

import (
	"fmt"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

// Waypoint symbols. Green marks a position where a lap begins, red where one
// ends and blue where one ends and the next begins.
const (
	symbolStart    = "Flag, Green"
	symbolFinish   = "Flag, Red"
	symbolBoundary = "Flag, Blue"
)

// Waypoints converts lap bounds to the waypoints marking where each lap begins
// and ends.
//
// Where a lap ends at the position the next begins, the two are emitted as a
// single waypoint: devices record the boundary trackpoint in both laps, so an
// uninterrupted recording would otherwise stack two waypoints on one position.
// Pausing the recording, starting a lap and resuming elsewhere moves the two
// apart, and they are then emitted separately.
func Waypoints(bounds []LapBounds) []gpxgo.GPXPoint {
	if len(bounds) == 0 {
		return nil
	}

	// each lap contributes one waypoint, plus one for the overall start; a lap
	// that does not begin where its predecessor ended contributes one more.
	waypoints := make([]gpxgo.GPXPoint, 0, len(bounds)+1)
	waypoints = append(waypoints, startWaypoint(bounds[0]))
	for i, lap := range bounds[:len(bounds)-1] {
		next := bounds[i+1]
		if samePosition(lap.Last, next.First) {
			waypoints = append(waypoints, boundaryWaypoint(lap, next))
			continue
		}
		waypoints = append(waypoints, finishWaypoint(lap), startWaypoint(next))
	}
	return append(waypoints, finishWaypoint(bounds[len(bounds)-1]))
}

// samePosition reports whether two points mark the same place. Elevation and
// time are deliberately not compared: the waypoints mark positions, so a lap
// boundary where the recording sat paused without moving is still one place.
func samePosition(a, b gpxgo.GPXPoint) bool {
	return a.Latitude == b.Latitude && a.Longitude == b.Longitude
}

func startWaypoint(lap LapBounds) gpxgo.GPXPoint {
	return waypoint(lap.First,
		fmt.Sprintf("Lap %d (start)", lap.Number),
		fmt.Sprintf("Start of lap %d", lap.Number),
		symbolStart)
}

func finishWaypoint(lap LapBounds) gpxgo.GPXPoint {
	return waypoint(lap.Last,
		fmt.Sprintf("Lap %d (finish)", lap.Number),
		fmt.Sprintf("End of lap %d", lap.Number),
		symbolFinish)
}

func boundaryWaypoint(lap, next LapBounds) gpxgo.GPXPoint {
	return waypoint(lap.Last,
		fmt.Sprintf("Lap %d (finish) / Lap %d (start)", lap.Number, next.Number),
		fmt.Sprintf("End of lap %d, start of lap %d", lap.Number, next.Number),
		symbolBoundary)
}

func waypoint(point gpxgo.GPXPoint, name, description, symbol string) gpxgo.GPXPoint {
	return gpxgo.GPXPoint{
		Point:       point.Point,
		Timestamp:   point.Timestamp,
		Name:        name,
		Description: description,
		Symbol:      symbol,
		Type:        "user",
	}
}
