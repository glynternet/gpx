package elevation

import (
	"fmt"

	"github.com/tkrajina/gpxgo/gpx"
)

type Profile []ProfilePoint

type ProfilePoint struct {
	Distance  float64 `json:"dist"`
	Elevation float64 `json:"ele"`
}

func CalculateProfile(points []gpx.GPXPoint) (Profile, error) {
	numPoints := len(points)
	if numPoints < 2 {
		return nil, fmt.Errorf("track must have at least 2 points: has %d", numPoints)
	}

	if points[0].Elevation.Null() {
		// TODO(glynternet): make so that null elevations up to first non-null elevation become overridden with
		//   the first non-null one.
		return nil, fmt.Errorf("first point elevation is null")
	}

	profile := Profile{
		{Distance: 0, Elevation: points[0].Elevation.Value()},
	}

	for i := 1; i < numPoints; i++ {
		prev := points[i-1]
		current := points[i]
		var ele float64
		if current.Elevation.Null() {
			// previous profile point elevation should always be present
			// because we check for first element being non-null then extrapolate
			ele = profile[i-1].Elevation
		} else {
			ele = current.Elevation.Value()
		}
		profile = append(profile, ProfilePoint{
			// TODO(glynternet): maybe distance 2D is actually what we want? How do other tracking applications work?
			Distance:  profile[i-1].Distance + prev.Distance3D(&current),
			Elevation: ele,
		})
	}
	return profile, nil
}
