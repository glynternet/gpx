package main

import (
	"io"
	"net/http"

	gpxhttp "github.com/glynternet/gpx/internal/http"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
)

func buildCmdTree(logger log.Logger, out io.Writer, rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:  "serve",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			address := ":8080"
			_ = logger.Log(log.Message("Listening"), log.KV{
				K: "address",
				V: address,
			})
			return (&http.Server{
				Addr:    address,
				Handler: http.HandlerFunc(gpxhttp.HandleElevation),
			}).ListenAndServe()
		},
	})
}
