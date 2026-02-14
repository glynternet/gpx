package validate

import (
	"fmt"

	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

type Validate[t any] func(t) error

func GPX(gpx *gpxgo.GPX, validations ...Validate[*gpxgo.GPX]) error {
	for _, validate := range validations {
		if err := validate(gpx); err != nil {
			return err
		}
	}
	return nil
}

func PopulatedTracks(validations ...Validate[gpxgo.GPXTrack]) Validate[*gpxgo.GPX] {
	return func(gpx *gpxgo.GPX) error {
		if len(gpx.Tracks) == 0 {
			return fmt.Errorf("gpx file must contain at least one track")
		}
		for i, track := range gpx.Tracks {
			for _, validate := range validations {
				if err := validate(track); err != nil {
					return fmt.Errorf("validating track (%d): %v", i, err)
				}
			}
		}
		return nil
	}
}

func SingleTrack(validations ...Validate[gpxgo.GPXTrack]) Validate[*gpxgo.GPX] {
	return func(gpx *gpxgo.GPX) error {
		if n := len(gpx.Tracks); n != 1 {
			return fmt.Errorf("gpx file must contain exactly 1 track but contains %d", n)
		}
		for _, validate := range validations {
			if err := validate(gpx.Tracks[0]); err != nil {
				return err
			}
		}
		return nil
	}
}

func SingleSegment() Validate[gpxgo.GPXTrack] {
	return func(track gpxgo.GPXTrack) error {
		if n := len(track.Segments); n != 1 {
			return fmt.Errorf("track must contain exactly 1 segment but contains %d", n)
		}
		return nil
	}
}
