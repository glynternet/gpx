package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/glynternet/gpx/pkg/tcx"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func tcxCmd(logger log.Logger, out io.Writer) *cobra.Command {
	tcxCmd := cobra.Command{
		Use: "tcx <tcx file>",
	}
	tcxCmd.AddCommand(tcxLapWaypointsCmd(logger, out))
	return &tcxCmd
}

func tcxLapWaypointsCmd(logger log.Logger, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "lap-waypoints <tcx-file>",
		Short: "Create a waypoint for the start and end of each lap in a TCX file.",
		Long: `Create a waypoint for the start and end of each lap in a TCX file.

Laps are read from TCX rather than GPX because Garmin Connect's GPX export
flattens an activity to a single track segment, discarding lap boundaries.

Where a lap ends at the position the next begins, the two are marked with a
single waypoint. Pausing the recording, starting a lap and resuming elsewhere
moves them apart, and they are then marked separately.

# usage
$ gpx tcx lap-waypoints activity.tcx
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fd, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer func() {
				// the file is only read, so a failure to close it cannot lose data
				_ = fd.Close()
			}()

			var db tcx.TrainingCenterDatabase
			if err := xml.NewDecoder(fd).Decode(&db); err != nil {
				return fmt.Errorf("decoding tcx content: %w", err)
			}

			bounds, skipped, err := db.LapBounds()
			if err != nil {
				return fmt.Errorf("reading lap bounds: %w", err)
			}
			for _, number := range skipped {
				_ = logger.Log(log.Message("Skipping lap that recorded no position"),
					log.KV{K: "lap", V: number})
			}
			if len(bounds) == 0 {
				if len(skipped) > 0 {
					return errors.New("no lap in the file recorded a position")
				}
				return errors.New("file contained no laps")
			}

			name := "lap waypoints"
			if len(db.Activities) > 0 && db.Activities[0].ID != "" {
				name = db.Activities[0].ID + " " + name
			}

			if err := gpxio.Write(out, gpxgo.GPX{
				Name:        name,
				Description: "Start and end waypoint markers for each lap in original file",
				Waypoints:   tcx.Waypoints(bounds),
			}); err != nil {
				return fmt.Errorf("writing gpx: %w", err)
			}
			return nil
		},
	}
}
