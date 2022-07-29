package gpx

import (
	"fmt"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func Split(gpx gpxgo.GPX) []gpxgo.GPX {
	var split []gpxgo.GPX
	for i, track := range gpx.Tracks {
		c := gpx
		if track.Name != "" {
			c.Name = track.Name
		} else {
			gpxName := gpx.Name
			if gpxName == "" {
				gpxName = "unknown"
			}
			name := fmt.Sprintf("%s-track-%d", gpxName, i+1)
			c.Name = name
			track.Name = name
		}
		// remove all extensions, they sometimes cause import issues and have never been useful
		track.Extensions = gpxgo.Extension{}
		for i := range track.Segments {
			// remove all extensions, they sometimes cause import issues and have never been useful
			track.Segments[i].Extensions = gpxgo.Extension{}
			for j := range track.Segments[i].Points {
				// remove all extensions, they sometimes cause import issues and have never been useful
				track.Segments[i].Points[j].Extensions = gpxgo.Extension{}
			}
		}
		c.Tracks = []gpxgo.GPXTrack{track}
		split = append(split, c)
	}
	return split
}
