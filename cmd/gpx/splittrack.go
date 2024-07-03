package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glynternet/pkg/log"

	"github.com/glynternet/gpx/pkg/gpx"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
)

func splitTrackCmd(logger log.Logger) *cobra.Command {
	var preOverlap float32
	cmd := cobra.Command{
		Use:   "split-track <gpx-file> <segment-count>", //TODO(glynternet): is this the right name?
		Short: "Split a GPX file containing a single track into multiple files, each with a subsection of the original track",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			segments, err := strconv.ParseUint(args[1], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid segment count: %s: %w", args[1], err)
			}
			path := args[0]
			content, err := gpxio.ReadFile(path)
			if err != nil {
				return err
			}
			if len(content.Tracks) == 0 {
				_ = log.Info(logger, log.Message("GPX file has no tracks, nothing to do"))
				return nil
			}
			if len(content.Tracks) > 1 {
				return fmt.Errorf("GPX file must contain exactly 1 track but contains %d", len(content.Tracks))
			}
			if len(content.Tracks[0].Segments) != 1 {
				return fmt.Errorf("GPX track must contain exactly 1 segment but contains %d", len(content.Tracks[0].Segments))
			}

			splitSegments, err := gpx.SplitPoints(content.Tracks[0].Segments[0].Points, uint(segments), preOverlap)
			if err != nil {
				return fmt.Errorf("could not split track points: %w", err)
			}

			var resolvePath func(i string) string
			if i := strings.LastIndexByte(path, '.'); i != -1 {
				basePath := path[:i]
				ext := path[i+1:]
				resolvePath = func(segIndex string) string {
					return basePath + "-" + segIndex + "." + ext
				}
			} else {
				resolvePath = func(segIndex string) string {
					return path + "-" + segIndex
				}
			}

			for i, points := range splitSegments {
				outFile := content
				iStr := strconv.Itoa(i)
				outFile.Name += "-" + iStr
				track := outFile.Tracks[0]
				if len(track.Name) != 0 {
					track.Name += "-" + iStr
				} else {
					track.Name = iStr
				}
				track.Segments[0].Points = points
				outPath := resolvePath(iStr)
				if err := writeSingleFile(outPath, *outFile); err != nil {
					return fmt.Errorf("could not write segment %d to file at %s: %w", i, outPath, err)
				}
				_ = log.Info(logger, log.Message("Split file written"), log.KV{K: "path", V: outPath}, log.KV{K: "track", V: outFile.Tracks[0].Name})
			}
			return nil
		},
	}
	cmd.Flags().Float32Var(&preOverlap, "pre-overlap-percent", 0, "percentage of segment size overlap to apply to start of each segment")
	return &cmd
}
