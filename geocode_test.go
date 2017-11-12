package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeocode(t *testing.T) {
	p, err := Geocode("Ada, Oklahoma, united states")
	assert.Nil(t, err)
	assert.Equal(t, 34.774531, p.Latitude)
	p, err = Geocode("edmonton, canada")
	assert.Nil(t, err)
	assert.Equal(t, 53.535411, p.Latitude)
	assert.Equal(t, -113.507996, p.Longitude)
}

func TestLocationFromIP(t *testing.T) {
	loc, err := LocationFromIP("198.199.67.130")
	assert.Nil(t, err)
	assert.Equal(t, "North Bergen, New Jersey, United States", loc)
}
