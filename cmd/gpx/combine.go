package main

import (
	"fmt"
	"io"

	gpxgo "github.com/tkrajina/gpxgo/gpx"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func combineCmd(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "combine [<gpx-file>...]",
		Short: "Combine multiple GPX into a single file containing all items.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var combined gpxgo.GPX
			for _, file := range args {
				content, err := gpxio.ReadFile(file)
				if err != nil {
					return err
				}
				combined.Routes = append(combined.Routes, content.Routes...)
				combined.Tracks = append(combined.Tracks, content.Tracks...)
				combined.Waypoints = append(combined.Waypoints, content.Waypoints...)
			}
			if err := gpxio.Write(out, combined); err != nil {
				return fmt.Errorf("error writing combined gpx: %w", err)
			}
			return nil
		},
	}
}
