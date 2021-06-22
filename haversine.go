package main

import "math"

// radius of the earth in miles.
const earthRadiusMi = 3958

// Distance calculates the shortest path between two coordinates on the surface
// of the Earth. This function returns two units of measure, the first is the
// distance in miles, the second is the distance in kilometers.
func HaversineDistance(p, q Coordinates) float64 {
	lat1 := p.Latitude * math.Pi / 180
	lon1 := p.Longitude * math.Pi / 180
	lat2 := q.Latitude * math.Pi / 180
	lon2 := q.Longitude * math.Pi / 180

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c * earthRadiusMi
}
