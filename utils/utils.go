package utils

import (
	"fmt"
	"math"
)

// HaversineDistance calcula la distancia entre dos puntos geográficos usando la fórmula haversine
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Radio de la Tierra en km

	// Convertir grados a radianes
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	// Aplicar fórmula haversine
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// DegreesToRadians convierte grados a radianes
func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// RadiansToDegrees convierte radianes a grados
func RadiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// ValidateCoordinates valida que las coordenadas estén en rangos válidos
func ValidateCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// CalculateBearing calcula el rumbo entre dos puntos geográficos
func CalculateBearing(lat1, lng1, lat2, lng2 float64) float64 {
	lat1Rad := DegreesToRadians(lat1)
	lat2Rad := DegreesToRadians(lat2)
	dLng := DegreesToRadians(lng2 - lng1)

	y := math.Sin(dLng) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLng)

	bearing := math.Atan2(y, x)
	bearing = RadiansToDegrees(bearing)

	// Normalizar a 0-360 grados
	if bearing < 0 {
		bearing += 360
	}

	return bearing
}

// CalculateMidpoint calcula el punto medio entre dos coordenadas
func CalculateMidpoint(lat1, lng1, lat2, lng2 float64) (float64, float64) {
	lat1Rad := DegreesToRadians(lat1)
	lng1Rad := DegreesToRadians(lng1)
	lat2Rad := DegreesToRadians(lat2)
	dLng := DegreesToRadians(lng2 - lng1)

	bx := math.Cos(lat2Rad) * math.Cos(dLng)
	by := math.Cos(lat2Rad) * math.Sin(dLng)

	midLat := math.Atan2(
		math.Sin(lat1Rad)+math.Sin(lat2Rad),
		math.Sqrt((math.Cos(lat1Rad)+bx)*(math.Cos(lat1Rad)+bx)+by*by))

	midLng := lng1Rad + math.Atan2(by, math.Cos(lat1Rad)+bx)

	return RadiansToDegrees(midLat), RadiansToDegrees(midLng)
}

// FormatDistance formatea la distancia en una cadena legible
func FormatDistance(km float64) string {
	if km < 1 {
		return fmt.Sprintf("%.0f m", km*1000)
	}
	return fmt.Sprintf("%.2f km", km)
}

// FormatDuration formatea la duración en una cadena legible
func FormatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%d min", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%d h", hours)
	}
	return fmt.Sprintf("%d h %d min", hours, mins)
}

// GetCardinalDirection convierte grados a dirección cardinal
func GetCardinalDirection(bearing float64) string {
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	index := int((bearing+22.5)/45) % 8
	return directions[index]
}
