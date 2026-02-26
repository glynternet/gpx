package elevation

import (
	"encoding/json"
	"fmt"
	"log"
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
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Distance   float64  `json:"dist"`
	Gain       float64  `json:"gain"`
	Loss       float64  `json:"loss"`
}

func CalculateProfiles(tracks []gpx.GPXTrack, waypoints []gpx.GPXPoint) ([]Profile, error) {
	trackPointWaypoints := map[gpx.Point][]waypointDetails{}
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
			categories := categoriesFromPoint(waypoint)
			if len(categories) == 0 && waypoint.Symbol != "" {
				categories = []string{waypoint.Symbol}
			}
			trackPointWaypoints[point] = append(trackPointWaypoints[point], waypointDetails{
				name:       waypoint.Name,
				categories: categories,
			})
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

		// it's possible for there to be multiple sequential points, all with the same lat/lon, which would produce duplicate
		// Waypoint values. So we need to deduplicate by hashing.
		// TODO: improve hashing function from json, which doesn't have ordered fields in the spec and so could, technically,
		//   produce inconsistent hashes
		var segmentWaypoints []Waypoint
		segmentWaypointsSeen := make(map[string]struct{})
		// checks if a given point at index i has any waypoints
		checkAndAppendWaypoints := func(i int, distance float64, cumulativeUpDown gpx.UphillDownhill) {
			trackPointWaypointNames, ok := trackPointWaypoints[points[i].Point]
			if !ok {
				return
			}
			for _, waypoint := range trackPointWaypointNames {
				wp := Waypoint{
					Name:       waypoint.name,
					Categories: waypoint.categories,
					Distance:   distance,
					Gain:       cumulativeUpDown.Uphill,
					Loss:       cumulativeUpDown.Downhill,
				}
				hash, err := json.Marshal(wp)
				if err != nil {
					log.Println("Could not marshal waypoint for hashing", err)
					return
				}
				hashStr := string(hash)
				if _, seen := segmentWaypointsSeen[hashStr]; seen {
					return
				}
				segmentWaypointsSeen[hashStr] = struct{}{}
				segmentWaypoints = append(segmentWaypoints, wp)
			}
		}
		var cumulativeUpDown gpx.UphillDownhill
		checkAndAppendWaypoints(0, 0, cumulativeUpDown)

		for i := 1; i < numPoints; i++ {
			prev := points[i-1]
			current := points[i]

			// previous profile point elevation should always be present
			// because we check for first element being non-null then extrapolate
			prevEle := trackPoints[i-1].Elevation
			var ele float64
			if current.Elevation.Null() {
				ele = prevEle
			} else {
				ele = current.Elevation.Value()
				if eleDelta := ele - prevEle; eleDelta > 0 {
					cumulativeUpDown.Uphill += eleDelta
				} else if eleDelta < 0 {
					cumulativeUpDown.Downhill -= eleDelta
				}
			}

			// TODO(glynternet): maybe distance 2D is actually what we want? How do other tracking applications work?
			dist := trackPoints[i-1].Distance + prev.Distance3D(&current)
			checkAndAppendWaypoints(i, dist, cumulativeUpDown)
			trackPoints = append(trackPoints, TrackPoint{
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

type waypointDetails struct {
	name       string
	categories []string
}

func categoriesFromPoint(p gpx.GPXPoint) []string {
	var categories []string
	for _, node := range p.Extensions.Nodes {
		if node.XMLName.Local == "Categories" {
			for _, child := range node.Nodes {
				if child.XMLName.Local == "Category" {
					categories = append(categories, child.Data)
				}
			}
		}
	}
	return categories
}
