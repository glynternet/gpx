package main

import (
	"io"

	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
)

func buildCmdTree(logger log.Logger, out io.Writer, rootCmd *cobra.Command) {
	rootCmd.AddCommand(combineCmd(out))
	rootCmd.AddCommand(csvCmd(out))
	rootCmd.AddCommand(geojsonCmd(logger, out))
	rootCmd.AddCommand(jsonCmd(logger, out))
	rootCmd.AddCommand(rotateCmd(out))
	rootCmd.AddCommand(separateCmd(logger))
	rootCmd.AddCommand(splitTrackCmd(logger))
	rootCmd.AddCommand(statCmd())
	rootCmd.AddCommand(tcxCmd(logger, out))
	rootCmd.AddCommand(timestampWaypointsCmd(out))
	rootCmd.AddCommand(trackWaypointsCmd(logger, out))
	rootCmd.AddCommand(trimCmd(logger))
}
