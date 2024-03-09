package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glynternet/pkg/log"
	"io"
	"os"
	"strconv"

	gpxio "github.com/glynternet/gpx/pkg/io"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func jsonCmd(logger log.Logger, out io.Writer) *cobra.Command {
	jsonCmd := cobra.Command{
		Use: "json <name> <csv file>",
	}
	jsonCmd.AddCommand(jsonWaypointsCmd(logger, out))
	return &jsonCmd
}

func jsonWaypointsCmd(logger log.Logger, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "waypoints <name> <json file>",
		Short: "Create gpx file from json file containing array of points.",
		Long:  "TODO",
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
			decoder := json.NewDecoder(fd)
			decoder.DisallowUnknownFields()

			type point struct {
				Name string `json:"name"`
				// matched lat lon names to GPX spec
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			}
			var ps []point
			if err := decoder.Decode(&ps); err != nil {
				return fmt.Errorf("docoding json content: %w", err)
			}
			if len(ps) == 0 {
				return errors.New("CSV contained no points")
			}

			gpxPs := make([]gpxgo.GPXPoint, len(ps))
			names := make(map[string]struct{})
			indexExtension := func(i int) string {
				return " (" + strconv.Itoa(i) + ")"
			}
			for i, p := range ps {
				resolvedName := p.Name
				for index := 1; ; index++ {
					checkName := p.Name
					if index > 1 {
						checkName += indexExtension(index)
					}
					if _, ok := names[checkName]; !ok {
						resolvedName = checkName
						names[checkName] = struct{}{}
						break
					}
				}
				if resolvedName != p.Name {
					if err := log.Warn(logger, log.Message("Duplicate name encountered, appending index"), log.KV{K: "name", V: p.Name}, log.KV{K: "renamed", V: resolvedName}); err != nil {
						panic(fmt.Errorf("error logging: %w", err))
					}
				}
				gpxPs[i] = gpxgo.GPXPoint{
					Point: gpxgo.Point{
						Latitude:  p.Lat,
						Longitude: p.Lon,
					},
					Name: resolvedName,
					Type: "user",
				}
			}

			return gpxio.Write(out, gpxgo.GPX{
				Name:      name,
				Waypoints: gpxPs,
			})
		},
	}
}
