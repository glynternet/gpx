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
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
}

type Waypoint struct {
	Distance float64 `json:"dist"`
	Name     string  `json:"name"`
}

func CalculateProfiles(tracks []gpx.GPXTrack, waypoints []gpx.GPXPoint) ([]Profile, error) {
	trackPointWaypoints := map[gpx.Point][]string{}
	for _, waypoint := range waypoints {
		minDistTrackIndex := -1
		var minDistPointIndex int
		minDist := math.MaxFloat64
		for trackIdx, track := range tracks {
			if len(track.Segments) == 0 {
				return nil, fmt.Errorf("track %d (%s) has no segments", trackIdx, track.Name)
			}
			if len(track.Segments) > 1 {
				return nil, fmt.Errorf("track %d (%s) has more than 1 segment, only 1 allowed per track", trackIdx, track.Name)
			}
			for pointIdx, point := range track.Segments[0].Points {
				if dist := point.Distance3D(&waypoint); minDistTrackIndex == -1 || dist < minDist {
					minDistTrackIndex = trackIdx
					minDistPointIndex = pointIdx
					minDist = dist
				}
			}
		}
		if minDistTrackIndex != -1 {
			point := tracks[minDistTrackIndex].Segments[0].Points[minDistPointIndex].Point
			trackPointWaypoints[point] = append(trackPointWaypoints[point], waypoint.Name)
		}
	}

	var profiles []Profile
	for _, track := range tracks {
		segment := track.Segments[0]
		points := segment.Points
		numPoints := len(points)
		if numPoints < 2 {
			return nil, fmt.Errorf("segments must have at least 2 points: has %d", numPoints)
		}

		if segment.Points[0].Elevation.Null() {
			// TODO(glynternet): make so that null elevations up to first non-null elevation become overridden with
			//   the first non-null one.
			return nil, fmt.Errorf("first point elevation is null")
		}
		trackPoints := []TrackPoint{
			{Distance: 0, Elevation: points[0].Elevation.Value(), Lat: points[0].Latitude, Lon: points[0].Longitude},
		}

		var segmentWaypoints []Waypoint
		// checks if a given point at index i has any waypoints
		checkAndAppendWaypoints := func(i int, distance float64) {
			trackPointWaypointNames, ok := trackPointWaypoints[points[i].Point]
			if !ok {
				return
			}
			for _, waypointName := range trackPointWaypointNames {
				segmentWaypoints = append(segmentWaypoints, Waypoint{
					Name:     waypointName,
					Distance: distance,
				})
			}
		}
		checkAndAppendWaypoints(0, 0)

		for i := 1; i < numPoints; i++ {
			prev := points[i-1]
			current := points[i]

			var ele float64
			if current.Elevation.Null() {
				// previous profile point elevation should always be present
				// because we check for first element being non-null then extrapolate
				ele = trackPoints[i-1].Elevation
			} else {
				ele = current.Elevation.Value()
			}

			dist := trackPoints[i-1].Distance + prev.Distance3D(&current)
			checkAndAppendWaypoints(i, dist)
			trackPoints = append(trackPoints, TrackPoint{
				// TODO(glynternet): maybe distance 2D is actually what we want? How do other tracking applications work?
				Distance:  dist,
				Elevation: ele,
				Lat:       current.Latitude,
				Lon:       current.Longitude,
			})
		}
		profiles = append(profiles, Profile{
			Track:     trackPoints,
			Waypoints: segmentWaypoints,
		})
	}

	return profiles, nil
}
