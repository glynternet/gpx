package main

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"

	gpxjson "github.com/glynternet/gpx/pkg/json"
	"github.com/stretchr/testify/require"
)

type response struct {
	Elements []interface{} `json:"elements"`
}

func TestWays(t *testing.T) {
	for i := 0; i < 6; i++ {
		path := "/home/g/tmp/pois/" + strconv.Itoa(i) + "-ways.json"
		f, err := os.Open(path)
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		var response response
		require.NoError(t, json.NewDecoder(f).Decode(&response), "path: %s", path)
		nodes := make(map[float64][2]float64)

		type way struct {
			nodes []float64
			tags  map[string]interface{}
		}
		var ways []way

		for _, element := range response.Elements {
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

		out, err := os.OpenFile("/home/g/tmp/pois/"+strconv.Itoa(i)+"-ways-gpx.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		require.NoError(t, err)
		defer func() { require.NoError(t, out.Close()) }()
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		require.NoError(t, encoder.Encode(pts))
	}
}

func TestNodes(t *testing.T) {
	for i := 0; i < 6; i++ {
		f, err := os.Open("/home/g/tmp/pois/" + strconv.Itoa(i) + "-nodes.json")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		var response response
		require.NoError(t, json.NewDecoder(f).Decode(&response))

		var pts []gpxjson.Point
		for _, element := range response.Elements {
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

		out, err := os.OpenFile("/home/g/tmp/pois/"+strconv.Itoa(i)+"-nodes-gpx.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		require.NoError(t, err)
		defer func() { require.NoError(t, out.Close()) }()
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		require.NoError(t, encoder.Encode(pts))
	}
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
	for _, tag := range []struct {
		key, value, symbol string
	}{
		{key: "leisure", value: "park", symbol: "Park"},
		{key: "amenity", value: "toilets", symbol: "Restroom"},
		{key: "amenity", value: "drinking_water", symbol: "Drinking Water"},
		{key: "natural", value: "peak", symbol: "Summit"},
		{key: "tourism", value: "viewpoint", symbol: "Scenic Area"},
		{key: "amenity", value: "bicycle_repair_station", symbol: "Mine"},
		{key: "amenity", value: "fast_food", symbol: "Fast Food"},
		{key: "amenity", value: "fuel", symbol: "Gas Station"},
	} {
		v, ok := tags[tag.key]
		if !ok {
			continue
		}
		if v != tag.value {
			continue
		}
		symbol = tag.symbol
		break
	}
	return symbol
}
