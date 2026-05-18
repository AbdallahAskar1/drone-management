package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHaversineMeters(t *testing.T) {
	// New York to London
	lat1, lng1 := 40.7128, -74.0060
	lat2, lng2 := 51.5074, -0.1278

	dist := HaversineMeters(lat1, lng1, lat2, lng2)

	// Should be approx 5570 km
	assert.InEpsilon(t, 5570000, dist, 0.01)

	// Same point
	assert.Equal(t, 0.0, HaversineMeters(lat1, lng1, lat1, lng1))
}

func TestETASeconds(t *testing.T) {
	// 1000 meters at 10 m/s
	eta := ETASeconds(0, 0, 0, 0.0089932, 10) // 0.0089932 degrees is ~1000m at equator
	assert.InDelta(t, 100, eta, 1)

	// Speed 0
	assert.Equal(t, int64(0), ETASeconds(0, 0, 1, 1, 0))
}
