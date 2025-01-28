package elevation

import (
	"fmt"
	"math"

	"github.com/tkrajina/gpxgo/gpx"
)

type Profile struct {
	Track     []TrackPoint `json:"track"`
	Waypoints []Waypoint   `json:"waypoints"`
}

type TrackPoint struct {
	Distance  float64 `json:"dist"`
	Elevation float64 `json:"ele"`
}

type Waypoint struct {
	Distance float64 `json:"dist"`
	Name     string  `json:"name"`
}

func CalculateProfile(track []gpx.GPXPoint, waypoints []gpx.GPXPoint) (Profile, error) {
	numPoints := len(track)
	if numPoints < 2 {
		return Profile{}, fmt.Errorf("track must have at least 2 points: has %d", numPoints)
	}

	if track[0].Elevation.Null() {
		// TODO(glynternet): make so that null elevations up to first non-null elevation become overridden with
		//   the first non-null one.
		return Profile{}, fmt.Errorf("first point elevation is null")
	}

	trackPointWaypoints := map[gpx.Point][]string{}
	for _, waypoint := range waypoints {
		minDistIndex := -1
		minDist := math.MaxFloat64
		for i, point := range track {
			if dist := point.Distance3D(&waypoint); minDistIndex == -1 || dist < minDist {
				minDist = dist
				minDistIndex = i
			}
		}
		if minDistIndex != -1 {
			trackPointWaypoints[track[minDistIndex].Point] = append(trackPointWaypoints[track[minDistIndex].Point], waypoint.Name)
		}
	}

	trackPoints := []TrackPoint{
		{Distance: 0, Elevation: track[0].Elevation.Value()},
	}

	var waypointDistances []Waypoint
	checkAndAppendDistances := func(point int, distance float64) {
		trackPointWaypointNames, ok := trackPointWaypoints[track[point].Point]
		if !ok {
			return
		}
		for _, waypointName := range trackPointWaypointNames {
			waypointDistances = append(waypointDistances, Waypoint{
				Name:     waypointName,
				Distance: distance,
			})
		}
	}
	checkAndAppendDistances(0, 0)

	for i := 1; i < numPoints; i++ {
		prev := track[i-1]
		current := track[i]

		var ele float64
		if current.Elevation.Null() {
			// previous profile point elevation should always be present
			// because we check for first element being non-null then extrapolate
			ele = trackPoints[i-1].Elevation
		} else {
			ele = current.Elevation.Value()
		}

		dist := trackPoints[i-1].Distance + prev.Distance3D(&current)
		checkAndAppendDistances(i, dist)
		trackPoints = append(trackPoints, TrackPoint{
			// TODO(glynternet): maybe distance 2D is actually what we want? How do other tracking applications work?
			Distance:  dist,
			Elevation: ele,
		})
	}

	return Profile{
		Track:     trackPoints,
		Waypoints: waypointDistances,
	}, nil
}
