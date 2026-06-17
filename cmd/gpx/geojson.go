package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/glynternet/gpx/pkg/geojson"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
)

func geojsonCmd(logger log.Logger, out io.Writer) *cobra.Command {
	geojsonCmd := cobra.Command{
		Use: "geojson <name> <geojson file>",
	}
	geojsonCmd.AddCommand(geojsonWaypointsCmd(logger, out))
	return &geojsonCmd
}

func geojsonWaypointsCmd(logger log.Logger, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "waypoints <name> <geojson file>",
		Short: "Create gpx file from a GeoJSON FeatureCollection of Point features.",
		Long: `Create gpx file from a GeoJSON (RFC 7946) FeatureCollection containing Point features.

Each feature's "name" property becomes the waypoint name, its coordinates
([lon, lat]) the position, its "category" property the waypoint symbol and its
"categories" property the waypoint categories. Unknown properties are ignored.

# usage
$ gpx geojson waypoints <gpx name> pois.geojson
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, file, err := waypointsArgs(args)
			if err != nil {
				return err
			}

			fd, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			// not disallowing unknown fields: GeoJSON properties carry rich
			// upstream data (e.g. raw OSM tags) that we intentionally ignore.
			var fc geojson.FeatureCollection
			if err := json.NewDecoder(fd).Decode(&fc); err != nil {
				return fmt.Errorf("decoding geojson content: %w", err)
			}

			ps, err := fc.Points()
			if err != nil {
				return fmt.Errorf("converting features to points: %w", err)
			}

			return writeWaypoints(logger, out, name, ps)
		},
	}
}
