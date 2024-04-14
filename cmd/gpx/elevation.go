package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/glynternet/gpx/pkg/gpx"
	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func elevationCmd(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "elevation <gpx-file>",
		Short: "Split a GPX into many files containing a single track each.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			profile, err := gpx.Profile(*content)
			if err != nil {
				return err
			}

			if err := json.NewEncoder(out).Encode(profile); err != nil {
				return fmt.Errorf("encoding profile: %w", err)
			}
			return nil
		},
	}
}
