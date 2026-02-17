package main

import (
	"fmt"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func statCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stat <gpx-file>",
		Short: "Show stats about a GPX file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			fmt.Println("Name:", content.Name)
			fmt.Println("Description:", content.Description)
			fmt.Println("Routes:", len(content.Routes))
			fmt.Println("Tracks:", len(content.Tracks))
			fmt.Println("Waypoints:", len(content.Waypoints))
			if len(content.Routes) > 0 {
				fmt.Println()
				for i, route := range content.Routes {
					fmt.Println("Route", i+1)
					fmt.Println("Name:", route.Name)
					fmt.Println("Description:", route.Description)
					fmt.Println("Points:", len(route.Points))
					fmt.Println("Length:", route.Length())
				}
			}
			if len(content.Tracks) > 0 {
				fmt.Println()
				for i, track := range content.Tracks {
					fmt.Println("Track", i+1)
					fmt.Println("Name:", track.Name)
					fmt.Println("Description:", track.Description)
					fmt.Println("Segments:", len(track.Segments))
					var length float64
					for _, segment := range track.Segments {
						length += segment.Length3D()
					}
					fmt.Println("Length (m):", length)
					upDown := track.UphillDownhill()
					fmt.Println("Gain (m):", upDown.Uphill)
					fmt.Println("Loss (m):", upDown.Downhill)
				}
			}
			if len(content.Waypoints) > 0 {
				fmt.Println()
				for i, waypoint := range content.Waypoints {
					fmt.Println("Waypoint", i+1)
					fmt.Println("Name:", waypoint.Name)
					fmt.Println("Description:", waypoint.Description)
				}
			}
			return nil
		},
	}
}
