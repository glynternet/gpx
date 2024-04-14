package gpx

import (
	"fmt"

	"github.com/glynternet/gpx/pkg/elevation"
	"github.com/tkrajina/gpxgo/gpx"
)

func Profile(content gpx.GPX) (elevation.Profile, error) {
	if len(content.Tracks) == 0 {
		return nil, fmt.Errorf("GPX file has no tracks")
	}
	// TODO(glynternet): could allow user to select track if there is more than one.
	//   or output for all of them?
	if len(content.Tracks) > 1 {
		return nil, fmt.Errorf("GPX file has more than one track")
	}

	if len(content.Tracks[0].Segments) == 0 {
		return nil, fmt.Errorf("track has no segments")
	}
	if len(content.Tracks[0].Segments) > 1 {
		return nil, fmt.Errorf("track has more than one segment")
	}

	profile, err := elevation.CalculateProfile(content.Tracks[0].Segments[0].Points)
	if err != nil {
		return nil, fmt.Errorf("calculating profile from points: %w", err)
	}
	return profile, nil
}
