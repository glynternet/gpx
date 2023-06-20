package main

import (
	"fmt"
	"io"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func trackWaypointsCmd(logger log.Logger, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "track-waypoints <gpx-file>",
		Short: "Create a waypoint for the start and end of each track in a GPX file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gpx, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			var points []gpxgo.GPXPoint
			for _, track := range gpx.Tracks {
				name := track.Name
				if len(track.Segments) == 0 {
					_ = logger.Log(log.Message("Skipping track with no segments"),
						log.KV{K: "track", V: name})
					continue
				}

				firstSegment := track.Segments[0]
				if len(firstSegment.Points) == 0 {
					_ = logger.Log(log.Message("Skipping track with empty first segment"),
						log.KV{K: "track", V: name})
					continue
				}

				lastSegment := track.Segments[len(track.Segments)-1]
				if len(lastSegment.Points) == 0 {
					_ = logger.Log(log.Message("Skipping track with empty first segment"),
						log.KV{K: "track", V: name})
					continue
				}

				firstPoint := firstSegment.Points[0]
				lastPoint := lastSegment.Points[len(lastSegment.Points)-1]
				points = append(points, gpxgo.GPXPoint{
					Point: gpxgo.Point{
						Latitude:  firstPoint.Latitude,
						Longitude: firstPoint.Longitude,
						Elevation: firstPoint.Elevation,
					},
					Name:        name + " (start)",
					Description: fmt.Sprintf("Start of track: %s", name),
					Symbol:      "Flag, Green",
					Type:        "user",
				}, gpxgo.GPXPoint{
					Point: gpxgo.Point{
						Latitude:  lastPoint.Latitude,
						Longitude: lastPoint.Longitude,
						Elevation: lastPoint.Elevation,
					},
					Name:        name + " (finish)",
					Description: fmt.Sprintf("End of track: %s", name),
					Symbol:      "Flag, Red",
					Type:        "user",
				})
			}

			return gpxio.Write(out, gpxgo.GPX{
				Name:        gpx.Name + " track waypoints",
				Description: "Start and end waypoint markers for each track in original file",
				Waypoints:   points,
			})
		},
	}
}
