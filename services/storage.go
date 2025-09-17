package services

import (
	"gowaze/models"
	"log"
	"sync"
	"time"
)

// Storage maneja el almacenamiento en memoria
type Storage struct {
	Users        map[int]*models.User
	Reports      map[int]*models.Report
	TrafficData  map[string]*models.TrafficData
	NextUserID   int
	NextReportID int
	mu           sync.RWMutex
}

// NewStorage crea una nueva instancia de Storage
func NewStorage() *Storage {
	return &Storage{
		Users:        make(map[int]*models.User),
		Reports:      make(map[int]*models.Report),
		TrafficData:  make(map[string]*models.TrafficData),
		NextUserID:   1,
		NextReportID: 1,
	}
}

// CreateUser crea o actualiza un usuario
func (s *Storage) CreateUser(username string, lat, lng float64) *models.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &models.User{
		ID:       s.NextUserID,
		Username: username,
		Lat:      lat,
		Lng:      lng,
		LastSeen: time.Now(),
	}
	s.Users[s.NextUserID] = user
	s.NextUserID++

	return user
}

// CreateReport crea un nuevo reporte
func (s *Storage) CreateReport(reportType string, lat, lng float64, description string, userID int) *models.Report {
	s.mu.Lock()
	defer s.mu.Unlock()

	report := &models.Report{
		ID:          s.NextReportID,
		Type:        reportType,
		Lat:         lat,
		Lng:         lng,
		Description: description,
		UserID:      userID,
		CreatedAt:   time.Now(),
		Votes:       1,
	}
	s.Reports[s.NextReportID] = report
	s.NextReportID++

	return report
}

// GetRecentReports obtiene reportes de las últimas 24 horas
func (s *Storage) GetRecentReports() []*models.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make([]*models.Report, 0, len(s.Reports))
	for _, report := range s.Reports {
		if time.Since(report.CreatedAt) < 24*time.Hour {
			reports = append(reports, report)
		}
	}
	return reports
}

// GetStats obtiene estadísticas generales
func (s *Storage) GetStats() (int, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usersOnline := len(s.Users)
	totalReports := len(s.Reports)
	trafficPoints := len(s.TrafficData)

	return usersOnline, totalReports, trafficPoints
}

// UpdateTrafficData actualiza datos de tráfico
func (s *Storage) UpdateTrafficData(key string, data *models.TrafficData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TrafficData[key] = data
}

// GetTrafficData obtiene todos los datos de tráfico
func (s *Storage) GetTrafficData() map[string]*models.TrafficData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Crear copia para evitar problemas de concurrencia
	data := make(map[string]*models.TrafficData)
	for k, v := range s.TrafficData {
		data[k] = v
	}
	return data
}

// InitSampleData inicializa datos de ejemplo
func (s *Storage) InitSampleData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reportes de ejemplo
	s.Reports[1] = &models.Report{
		ID:          1,
		Type:        "traffic",
		Lat:         14.0818,
		Lng:         -87.2068,
		Description: "Tráfico pesado en el centro de San Pedro Sula",
		UserID:      1,
		CreatedAt:   time.Now().Add(-10 * time.Minute),
		Votes:       5,
	}

	s.Reports[2] = &models.Report{
		ID:          2,
		Type:        "police",
		Lat:         14.0900,
		Lng:         -87.2100,
		Description: "Control policial en Bulevar del Norte",
		UserID:      1,
		CreatedAt:   time.Now().Add(-5 * time.Minute),
		Votes:       3,
	}

	s.Reports[3] = &models.Report{
		ID:          3,
		Type:        "accident",
		Lat:         14.0750,
		Lng:         -87.2200,
		Description: "Accidente menor en intersección",
		UserID:      1,
		CreatedAt:   time.Now().Add(-15 * time.Minute),
		Votes:       7,
	}

	s.NextReportID = 4

	log.Println("✅ Datos de ejemplo inicializados")
}

// StartCleanup inicia la limpieza automática de datos antiguos
func (s *Storage) StartCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup limpia datos antiguos
func (s *Storage) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Limpiar usuarios inactivos (más de 1 hora)
	for id, user := range s.Users {
		if time.Since(user.LastSeen) > time.Hour {
			delete(s.Users, id)
		}
	}

	// Limpiar reportes antiguos (más de 24 horas)
	for id, report := range s.Reports {
		if time.Since(report.CreatedAt) > 24*time.Hour {
			delete(s.Reports, id)
		}
	}

	// Limpiar datos de tráfico antiguos (más de 1 hora)
	for key, traffic := range s.TrafficData {
		if time.Since(traffic.Timestamp) > time.Hour {
			delete(s.TrafficData, key)
		}
	}

	log.Printf("🧹 Limpieza automática completada. Users: %d, Reports: %d, Traffic: %d",
		len(s.Users), len(s.Reports), len(s.TrafficData))
}