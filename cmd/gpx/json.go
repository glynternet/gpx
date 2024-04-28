package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	gpxio "github.com/glynternet/gpx/pkg/io"
	json2 "github.com/glynternet/gpx/pkg/json"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
	gpxgo "github.com/tkrajina/gpxgo/gpx"
)

func jsonCmd(logger log.Logger, out io.Writer) *cobra.Command {
	jsonCmd := cobra.Command{
		Use: "json <name> <json file>",
	}
	jsonCmd.AddCommand(jsonWaypointsCmd(logger, out))
	return &jsonCmd
}

func jsonWaypointsCmd(logger log.Logger, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "waypoints <name> <json file>",
		Short: "Create gpx file from json file containing array of points.",
		Long: `Create gpx file from json file containing array of points of name, lat, lon fields.

e.g.

# points.json file content
[
	{"name":"point a","lat":1.23,"lon":4.56},
	{"name":"point b","lat":7.89,"lon":0.12}
]

# usage
$ gpx json <gpx name> points.json
`,
		Args: cobra.ExactArgs(2),
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

			var ps []json2.Point
			if err := decoder.Decode(&ps); err != nil {
				return fmt.Errorf("docoding json content: %w", err)
			}
			if len(ps) == 0 {
				return errors.New("json file contained no points")
			}

			gpxPs := make([]gpxgo.GPXPoint, len(ps))
			names := make(map[string]struct{})
			indexExtension := func(i int) string {
				return " (" + strconv.Itoa(i) + ")"
			}
			for i, p := range ps {
				if p.Name == "" {
					return errors.New("point with empty name")
				}
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
					Name:        resolvedName,
					Description: p.Description,
					Type:        "user",
					Symbol:      p.Symbol,
				}
			}

			return gpxio.Write(out, gpxgo.GPX{
				Name:      name,
				Waypoints: gpxPs,
			})
		},
	}
}
