package services

import (
	"fmt"
	"gowaze/models"
	"time"
)

// TrafficService simula datos de tráfico en tiempo real
type TrafficService struct {
	storage *Storage
}

// NewTrafficService crea una nueva instancia del servicio de tráfico
func NewTrafficService(storage *Storage) *TrafficService {
	return &TrafficService{
		storage: storage,
	}
}

// Start inicia el simulador de tráfico
func (ts *TrafficService) Start() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ts.simulateTrafficData()
	}
}

// simulateTrafficData simula datos de tráfico para diferentes zonas
func (ts *TrafficService) simulateTrafficData() {
	// Zonas de San Pedro Sula para simular tráfico
	locations := []models.Location{
		{14.0818, -87.2068}, // Centro - Plaza Central
		{14.0900, -87.2100}, // Zona Norte - Bulevar
		{14.0700, -87.2000}, // Zona Sur
		{14.0800, -87.1900}, // Zona Este
		{14.0750, -87.2200}, // Zona Oeste
		{14.0950, -87.2150}, // Universidad UNAH
		{14.0650, -87.2050}, // Hospital San Felipe
		{14.0850, -87.1950}, // Mall Multiplaza
	}

	for i, loc := range locations {
		key := fmt.Sprintf("%.4f,%.4f", loc.Lat, loc.Lng)

		// Simular velocidad basada en hora del día y ubicación
		speed := ts.calculateSpeed(i, loc)
		congestion := ts.getCongestionLevel(speed)

		trafficData := &models.TrafficData{
			Lat:        loc.Lat,
			Lng:        loc.Lng,
			Speed:      speed,
			Congestion: congestion,
			Timestamp:  time.Now(),
		}

		ts.storage.UpdateTrafficData(key, trafficData)
	}
}

// calculateSpeed calcula la velocidad basada en diferentes factores
func (ts *TrafficService) calculateSpeed(locationIndex int, loc models.Location) float64 {
	hour := time.Now().Hour()
	baseSpeed := 45.0

	// Horas pico: reducir velocidad
	if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
		baseSpeed = 25.0
	}

	// Velocidad nocturna más alta
	if hour >= 22 || hour <= 5 {
		baseSpeed = 55.0
	}

	// Variabilidad por ubicación
	locationVariation := float64((locationIndex*7 + int(time.Now().Unix())) % 20 - 10)
	speed := baseSpeed + locationVariation

	// Asegurar velocidad mínima
	if speed < 5 {
		speed = 5
	}

	// Velocidad máxima urbana
	if speed > 70 {
		speed = 70
	}

	return speed
}

// getCongestionLevel determina el nivel de congestión basado en velocidad
func (ts *TrafficService) getCongestionLevel(speed float64) string {
	if speed > 40 {
		return "low"
	} else if speed > 25 {
		return "medium"
	}
	return "high"
}

// GetCurrentTrafficSummary obtiene un resumen del tráfico actual
func (ts *TrafficService) GetCurrentTrafficSummary() map[string]int {
	trafficData := ts.storage.GetTrafficData()
	summary := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	for _, data := range trafficData {
		// Solo contar datos recientes (menos de 1 hora)
		if time.Since(data.Timestamp) < time.Hour {
			summary[data.Congestion]++
		}
	}

	return summary
}