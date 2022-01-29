package main

import (
	"fmt"
	"io"
	"strconv"

	"github.com/glynternet/gpx/pkg/gpx"
	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func rotateCmd(out io.Writer) *cobra.Command {
	var rotation int
	return &cobra.Command{
		Use:   "rotate <gpx-file> <points>",
		Short: "Rotate a single track GPX file by a given number of points.",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}
			validRotation, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("points arg must be integer: %w", err)
			}
			rotation = validRotation
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			if err := validate(content); err != nil {
				return fmt.Errorf("content invalid or not supported: %w", err)
			}

			content.Tracks[0].Segments[0].Points = gpx.Rotated(content.Tracks[0].Segments[0].Points, rotation)
			return gpxio.Write(out, *content)
		},
	}
}
