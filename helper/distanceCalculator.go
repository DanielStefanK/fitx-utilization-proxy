package helper

import "math"

// CalculateDistance: uses a special formula with the earth radius (6371 km)
func CalculateDistance(lat1, long1 float64, lat2, long2 float64) float64 {
	return math.Acos(
		math.Sin(lat1)*math.Sin(lat2)+math.Cos(lat1)*math.Cos(lat2)*math.Cos(long2-long1)) * 6371
}
