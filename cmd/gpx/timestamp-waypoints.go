package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/glynternet/gpx/pkg/gpx"
	"github.com/glynternet/gpx/pkg/gpx/validate"
	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func timestampWaypointsCmd(out io.Writer) *cobra.Command {
	var waypointName string
	command := cobra.Command{
		Use:   "timestamp-waypoints <gpx-file> [<unix-timestamp>...]",
		Short: "Create a waypoint at the recorded location for each of the given unix timestamps.",
		Long: "Create a waypoint at the recorded location for each of the given unix timestamps.\n\n" +
			"Locations are interpolated between the recorded points either side of each timestamp.\n" +
			"Timestamps are read one-per-line from stdin when none are given as arguments.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := gpxio.ReadFile(args[0])
			if err != nil {
				return err
			}

			if err := validate.GPX(content, validate.SingleTrack(validate.SingleSegment())); err != nil {
				return fmt.Errorf("validating gpx: %w", err)
			}

			var times []time.Time
			if len(args) > 1 {
				times, err = parseUnixTimestamps(args[1:])
			} else {
				times, err = readUnixTimestamps(cmd.InOrStdin())
			}
			if err != nil {
				return err
			}
			if len(times) == 0 {
				return errors.New("at least one unix timestamp must be provided")
			}

			track := content.Tracks[0]
			points, err := gpx.PointsAt(track.Segments[0].Points, times)
			if err != nil {
				return fmt.Errorf("getting points at given timestamps: %w", err)
			}

			waypoints := make([]gpxgo.GPXPoint, len(points))
			for i, point := range points {
				timestamp := strconv.FormatInt(point.Timestamp.Unix(), 10)
				// without a given name, the timestamp alone is all there is to identify a waypoint by
				pointName := timestamp
				if waypointName != "" {
					pointName = fmt.Sprintf("%s %d (%s)", waypointName, i+1, timestamp)
				}
				waypoints[i] = gpxgo.GPXPoint{
					Point:       point.Point,
					Timestamp:   point.Timestamp,
					Name:        pointName,
					Description: fmt.Sprintf("Location of track %q at %s", track.Name, point.Timestamp.UTC().Format(time.RFC3339)),
					Symbol:      "Flag, Blue",
					Type:        "user",
				}
			}

			name := "timestamp waypoints"
			if content.Name != "" {
				name = content.Name + " " + name
			}

			if err := gpxio.Write(out, gpxgo.GPX{
				Name:        name,
				Description: "Interpolated locations of the original track at the given timestamps",
				Waypoints:   waypoints,
			}); err != nil {
				return fmt.Errorf("writing gpx: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&waypointName, "name", "",
		"name to give each waypoint, suffixed with an incrementing index and the waypoint's timestamp. Defaults to the timestamp alone")
	return &command
}

func readUnixTimestamps(r io.Reader) ([]time.Time, error) {
	var values []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if value := strings.TrimSpace(scanner.Text()); value != "" {
			values = append(values, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading timestamps: %w", err)
	}
	return parseUnixTimestamps(values)
}

func parseUnixTimestamps(values []string) ([]time.Time, error) {
	times := make([]time.Time, len(values))
	for i, value := range values {
		seconds, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing unix timestamp %q: %w", value, err)
		}
		times[i] = time.Unix(seconds, 0).UTC()
	}
	return times, nil
}
