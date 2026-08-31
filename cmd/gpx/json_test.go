package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	gpxjson "github.com/glynternet/gpx/pkg/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata rather than asserting against them")

type response struct {
	Elements []interface{} `json:"elements"`
}

func TestWays(t *testing.T) {
	resp := decodeOverpassResponse(t, "ways.json")
	nodes := make(map[float64][2]float64)

	type way struct {
		nodes []float64
		tags  map[string]interface{}
	}
	var ways []way

	for _, element := range resp.Elements {
		elementMetadata, ok := element.(map[string]interface{})
		require.True(t, ok)
		switch elementMetadata["type"] {
		case "node":
			nodes[elementMetadata["id"].(float64)] = [2]float64{elementMetadata["lat"].(float64), elementMetadata["lon"].(float64)}
		case "way":
			var nodes []float64
			for _, node := range elementMetadata["nodes"].([]interface{}) {
				nodes = append(nodes, node.(float64))
			}
			ways = append(ways, way{
				nodes: nodes,
				tags:  elementMetadata["tags"].(map[string]interface{}),
			})
		}
	}

	var pts []gpxjson.Point
	for _, way := range ways {
		var lat, lon float64
		nodesLen := float64(len(way.nodes))
		for _, nodeID := range way.nodes {
			lat += nodes[nodeID][0] / nodesLen
			lon += nodes[nodeID][1] / nodesLen
		}

		name, err := resolveName(way.tags)
		require.NoError(t, err)

		desc, err := json.Marshal(way.tags)
		require.NoError(t, err)

		pts = append(pts, gpxjson.Point{
			Name:        name,
			Lat:         lat,
			Lon:         lon,
			Description: string(desc),
			Symbol:      resolveSymbol(way.tags),
		})
	}

	assertGoldenPoints(t, "ways-gpx.json", pts)
}

func TestNodes(t *testing.T) {
	resp := decodeOverpassResponse(t, "nodes.json")

	var pts []gpxjson.Point
	for _, element := range resp.Elements {
		elementMetadata, ok := element.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "node", elementMetadata["type"])

		tags := elementMetadata["tags"].(map[string]interface{})

		name, err := resolveName(tags)
		require.NoError(t, err)

		desc, err := json.Marshal(tags)
		require.NoError(t, err)

		pts = append(pts, gpxjson.Point{
			Name:        name,
			Lat:         elementMetadata["lat"].(float64),
			Lon:         elementMetadata["lon"].(float64),
			Description: string(desc),
			Symbol:      resolveSymbol(tags),
		})
	}

	assertGoldenPoints(t, "nodes-gpx.json", pts)
}

// decodeOverpassResponse decodes the named Overpass API response from testdata.
func decodeOverpassResponse(t *testing.T, name string) response {
	t.Helper()
	path := filepath.Join("testdata", name)
	f, err := os.Open(path)
	require.NoError(t, err, "path: %s", path)
	t.Cleanup(func() { require.NoError(t, f.Close()) })

	var resp response
	require.NoError(t, json.NewDecoder(f).Decode(&resp), "path: %s", path)
	return resp
}

// assertGoldenPoints asserts that pts encode to the content of the named
// golden file in testdata, rewriting that file instead when -update is passed.
func assertGoldenPoints(t *testing.T, name string, pts []gpxjson.Point) {
	t.Helper()
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(pts))

	path := filepath.Join("testdata", name)
	if *update {
		require.NoError(t, os.WriteFile(path, buf.Bytes(), 0600))
		return
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err, "path: %s (run with -update to create it)", path)
	assert.Equal(t, string(want), buf.String())
}

func resolveName(tags map[string]interface{}) (string, error) {
	if n, ok := tags["name"]; ok {
		return n.(string), nil
	} else if a, ok := tags["amenity"]; ok {
		return a.(string), nil
	} else if a, ok := tags["tourism"]; ok {
		return a.(string), nil
	} else if a, ok := tags["leisure"]; ok {
		return a.(string), nil
	}
	return "", errors.New("no suitable tag for name")
}

func resolveSymbol(tags map[string]interface{}) string {
	var symbol string
	for _, symbolMatchers := range []struct {
		tags   map[string]string
		symbol string
	}{
		{tags: map[string]string{"leisure": "park"}, symbol: "Park"},
		{tags: map[string]string{"amenity": "toilets"}, symbol: "Restroom"},
		{tags: map[string]string{"amenity": "drinking_water"}, symbol: "Drinking Water"},
		{tags: map[string]string{"natural": "peak"}, symbol: "Summit"},
		{tags: map[string]string{"tourism": "viewpoint"}, symbol: "Scenic Area"},
		{tags: map[string]string{"amenity": "bicycle_repair_station"}, symbol: "Mine"},
		{tags: map[string]string{"amenity": "fast_food"}, symbol: "Fast Food"},
		{tags: map[string]string{"amenity": "fuel"}, symbol: "Gas Station"},
		{tags: map[string]string{"amenity": "pub"}, symbol: "Bar"},
		{tags: map[string]string{"amenity": "cafe"}, symbol: "Restaurant"},
		{tags: map[string]string{"tourism": "picnic_site"}, symbol: "Picnic Area"},
		{tags: map[string]string{"amenity": "restaurant", "cuisine": "pizza"}, symbol: "Pizza"},
		{tags: map[string]string{"amenity": "restaurant"}, symbol: "Restaurant"},
		{tags: map[string]string{"amenity": "ice_cream"}, symbol: "Fast Food"},
		{tags: map[string]string{"tourism": "camp_pitch"}, symbol: "Campground"},
		{tags: map[string]string{"leisure": "nature_reserve"}, symbol: "Park"},
		{tags: map[string]string{"amenity": "shelter"}, symbol: "Building"},
		{tags: map[string]string{"amenity": "place_of_worship"}, symbol: "Church"},
	} {
		match := true
		for k, matcherV := range symbolMatchers.tags {
			v, ok := tags[k]
			if !ok || v != matcherV {
				match = false
				break
			}
		}
		if match {
			symbol = symbolMatchers.symbol
			break
		}
	}
	return symbol
}

func TestExampleJSONPoints(t *testing.T) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	err := encoder.Encode([]gpxjson.Point{{
		Name:        "Foo",
		Lat:         1,
		Lon:         2,
		Description: "Foo is a fantastic bar",
		Symbol:      "shrug",
	}, {
		Name: "Baz",
		Lat:  -1,
		Lon:  -2,
	}})
	require.NoError(t, err)
	assert.Equal(t, exampleJsonPoints, buf.String())
}
