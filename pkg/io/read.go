package io

import (
	"fmt"
	"github.com/tkrajina/gpxgo/gpx"
	"os"
)

func ReadFile(path string) (*gpx.GPX, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading content file:%q: %w", path, err)
	}

	gpxData, err := gpx.Parse(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("closing file:%q: %w", path, err)
	}
	return gpxData, nil
}
