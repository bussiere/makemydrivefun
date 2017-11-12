package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFindNearbyFeatures(b *testing.B) {
	r, _ := New(portland, beaverton)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.FindNearbyFeatures()
	}
}

func TestRoadTrip(t *testing.T) {
	r, err := New(portland, beaverton)
	assert.Nil(t, err)
	r.FindNearbyFeatures()
	r.Dump("output.json")
}
