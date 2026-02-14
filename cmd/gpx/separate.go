package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glynternet/gpx/pkg/gpx/validate"
	"github.com/glynternet/pkg/log"

	gpxgo "github.com/tkrajina/gpxgo/gpx"

	"github.com/glynternet/gpx/pkg/gpx"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func separateCmd(logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "separate <gpx-file>",
		Short: "Separate a GPX into many files containing a single track each.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			if err := validate.GPX(content, validate.PopulatedTracks(validate.SingleSegment())); err != nil {
				return fmt.Errorf("validating gpx: %w", err)
			}

			for _, gpx := range gpx.Split(*content) {
				path := filepath.Join(".", strings.ReplaceAll(gpx.Name, string([]byte{os.PathSeparator}), `-`)+".gpx")
				if err := writeSingleFile(path, gpx); err != nil {
					return fmt.Errorf("writing gpx to file: %w", err)
				}
				_ = log.Info(logger, log.Message("Split file written"), log.KV{K: "path", V: path}, log.KV{K: "track", V: gpx.Tracks[0].Name})

			}
			return nil
		},
	}
}

func writeSingleFile(path string, gpx gpxgo.GPX) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	if err := gpxio.Write(file, gpx); err != nil {
		_ = file.Close()
		return fmt.Errorf("writing to file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing file: %w", err)
	}
	return nil
}
