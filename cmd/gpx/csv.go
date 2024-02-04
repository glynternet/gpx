package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func csvCmd(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "csv <name> <csv file>",
		Short: "Create gpx file from csv with line of lat,lot,ele",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return errors.New("name must not be empty")
			}
			file := args[1]
			if file == "" {
				return errors.New("file must not be empty")
			}

			fd, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			r := csv.NewReader(fd)
			r.FieldsPerRecord = 3
			var points []gpxgo.GPXPoint
			for i := 1; ; i++ {
				entry, err := r.Read()
				if err != nil {
					if err != io.EOF {
						return fmt.Errorf("reading entry (line %d): %w", i, err)
					}
					break
				}
				lat, err := strconv.ParseFloat(entry[0], 64)
				if err != nil {
					return fmt.Errorf("parsing latitude field (line %d): %w", i, err)
				}
				lon, err := strconv.ParseFloat(entry[1], 64)
				if err != nil {
					return fmt.Errorf("parsing longitude field (line %d): %w", i, err)
				}
				ele, err := strconv.ParseFloat(entry[2], 64)
				if err != nil {
					return fmt.Errorf("parsing elevation field (line %d): %w", i, err)
				}
				points = append(points, gpxgo.GPXPoint{
					Point: gpxgo.Point{
						Latitude:  lat,
						Longitude: lon,
						Elevation: *gpxgo.NewNullableFloat64(ele),
					},
				})
			}
			if len(points) == 0 {
				return errors.New("CSV contained no points")
			}

			if err := gpxio.Write(out, gpxgo.GPX{
				Name: name,
				Tracks: []gpxgo.GPXTrack{{
					Name: name,
					Segments: []gpxgo.GPXTrackSegment{{
						Points: points,
					}},
				}},
			}); err != nil {
				return fmt.Errorf("writing gpx file: %w", err)
			}

			return nil
		},
	}
}
