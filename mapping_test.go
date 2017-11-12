package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var edmonton = Point{-113.4909, 54.5444}
var beaverton = Point{-122.83819, 45.5001634}
var portland = Point{-122.6765, 45.5231}

func TestRoute1(t *testing.T) {
	route, err := portland.DrivingRouteTo(beaverton)
	assert.Nil(t, err)
	assert.Equal(t, 63, len(route.Steps()))
	assert.Equal(t, 18, int(route.Duration()))
	assert.Equal(t, 10, int(route.Distance()))
	assert.Equal(t, 12, int(portland.DistanceAsCrowFlies(beaverton)))
}

func TestNearestCity(t *testing.T) {
	assert.Equal(t, "Aloha, Oregon, United States", beaverton.NearestCityWithin(5))
}

func BenchmarkNearestCity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		beaverton.NearestCityWithin(5)
	}
}
