package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadGeoJSON(t *testing.T) {
	g, err := LoadGeoJSONFile("static/data/features.geojson")
	assert.Nil(t, err)
	assert.Equal(t, len(g.Features), 17183)
	var features []GeoJSONPointsFeature
	features = g.Features
	assert.Equal(t, len(features), 17183)
}
