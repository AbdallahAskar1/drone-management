package utils

import "math"

const earthRadiusM = 6371000.0

func HaversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLng := rad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func ETASeconds(fromLat, fromLng, toLat, toLng, speedMS float64) int64 {
	if speedMS <= 0 {
		return 0
	}
	d := HaversineMeters(fromLat, fromLng, toLat, toLng)
	return int64(math.Round(d / speedMS))
}
