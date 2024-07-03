package io

import (
	"fmt"
	"os"

	"github.com/tkrajina/gpxgo/gpx"
)

func ReadFile(path string) (*gpx.GPX, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading content file:%q: %w", path, err)
	}

	data, err := gpx.ParseBytes(content)
	if err != nil {
		return nil, fmt.Errorf("parsing content data: %w", err)
	}
	return data, nil
}
