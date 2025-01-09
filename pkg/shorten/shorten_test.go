package shorten_test

import (
	"github.com/glynternet/gpx/pkg/shorten"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDevowel(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  string
		skip   uint
		output string
	}{{
		name: "empty returns empty",
	}, {
		name:   "no vowels returns same",
		input:  "bcd",
		output: "bcd",
	}, {
		name:   "only vowels returns empty",
		input:  "aeiou",
		output: "aeiou",
	}, {
		name:   "strips all vowels when no skip",
		input:  "abcdefghijklmnopqrstuvwxyz",
		output: "bcdfghjklmnpqrstvwxyz",
	}, {
		name:   "strips vowels of all cases",
		input:  "abcdefGhIjklmnOpqrstUvwxyz",
		output: "bcdfGhjklmnpqrstvwxyz",
	}, {
		name:   "skips stripping vowels when skip configured",
		input:  "abcdefGhIjklmnOpqrstUvwxyz",
		output: "abcdefGhIjklmnpqrstvwxyz",
		skip:   3,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.output, shorten.Devowel(tc.skip)(tc.input))
		})
	}
}
