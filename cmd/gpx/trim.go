package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/glynternet/gpx/pkg/gpx/validate"
	"github.com/glynternet/pkg/log"

	"github.com/glynternet/gpx/pkg/gpx"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func trimCmd(logger log.Logger) *cobra.Command {
	var trimStart float64
	var trimEnd float64
	command := cobra.Command{
		Use:   "trim <gpx-file>",
		Short: "Trim the start and/or end of a GPX track.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			if trimStart == 0 && trimEnd == 0 {
				return errors.New("start or end trim distance must be provided")
			}
			if trimStart < 0 {
				return errors.New("start trim distance cannot be negative")
			}
			if trimEnd < 0 {
				return errors.New("end trim distance cannot be negative")
			}

			if err := validate.GPX(content, validate.SingleTrack(validate.SingleSegment())); err != nil {
				return fmt.Errorf("validating gpx: %w", err)
			}

			pts := content.Tracks[0].Segments[0].Points
			if trimStart > 0 {
				index, err := gpx.DistanceIndex(pts, trimStart)
				if err != nil {
					return fmt.Errorf("trimming start: %w", err)
				}
				pts = pts[index+1:]
			}
			if trimEnd > 0 {
				slices.Reverse(pts)
				index, err := gpx.DistanceIndex(pts, trimEnd)
				if err != nil {
					return fmt.Errorf("trimming end: %w", err)
				}
				pts = pts[index+1:]
				slices.Reverse(pts)
			}
			content.Tracks[0].Segments[0].Points = pts
			if err := gpxio.Write(os.Stdout, *content); err != nil {
				return fmt.Errorf("writing gpx to stdout: %w", err)
			}
			return nil
		},
	}
	command.Flags().Float64Var(&trimStart, "start", 0.0, "Distance (m) to trim from start of track")
	command.Flags().Float64Var(&trimEnd, "end", 0.0, "Distance (m) to trim from end of track")
	return &command
}
