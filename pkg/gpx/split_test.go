package gpx_test

import (
	"testing"

	"github.com/glynternet/gpx/pkg/gpx"
	"github.com/stretchr/testify/assert"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func TestSplit(t *testing.T) {
	t.Run("zero tracks returns empty slice", func(t *testing.T) {
		in := gpxgo.GPX{
			Name:   "foo",
			Tracks: nil,
		}
		out := gpx.Split(in)
		assert.Equal(t, []gpxgo.GPX(nil), out)
	})

	t.Run("single track returns slice of same file", func(t *testing.T) {
		in := gpxgo.GPX{
			Name:   "foo",
			Tracks: []gpxgo.GPXTrack{{Name: "bar"}},
		}
		out := gpx.Split(in)
		assert.Equal(t, []gpxgo.GPX{{Name: "bar", Tracks: []gpxgo.GPXTrack{{Name: "bar"}}}}, out)
	})

	t.Run("multiple tracks returns slice of individual tracks parent GPX named to match track", func(t *testing.T) {
		in := gpxgo.GPX{
			Name:   "foo",
			Tracks: []gpxgo.GPXTrack{{Name: "bar"}, {Name: "baz"}},
		}
		out := gpx.Split(in)
		assert.Equal(t, []gpxgo.GPX{
			{
				Name:   "bar",
				Tracks: []gpxgo.GPXTrack{{Name: "bar"}},
			}, {
				Name:   "baz",
				Tracks: []gpxgo.GPXTrack{{Name: "baz"}},
			},
		}, out)
	})

	t.Run("tracks with missing names are given name based on parent file name with track index", func(t *testing.T) {
		in := gpxgo.GPX{
			Name:   "foo",
			Tracks: []gpxgo.GPXTrack{{}, {}},
		}
		out := gpx.Split(in)
		assert.Equal(t, []gpxgo.GPX{
			{
				Name:   "foo-track-1",
				Tracks: []gpxgo.GPXTrack{{Name: "foo-track-1"}},
			}, {
				Name:   "foo-track-2",
				Tracks: []gpxgo.GPXTrack{{Name: "foo-track-2"}},
			},
		}, out)
	})

	t.Run("parent gpx missing name results in \"unknown\" name", func(t *testing.T) {
		in := gpxgo.GPX{
			Tracks: []gpxgo.GPXTrack{{}, {}},
		}
		out := gpx.Split(in)
		assert.Equal(t, []gpxgo.GPX{
			{
				Name:   "unknown-track-1",
				Tracks: []gpxgo.GPXTrack{{Name: "unknown-track-1"}},
			}, {
				Name:   "unknown-track-2",
				Tracks: []gpxgo.GPXTrack{{Name: "unknown-track-2"}},
			},
		}, out)
	})
}
