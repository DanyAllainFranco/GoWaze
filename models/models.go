package models

import "time"

// User representa un usuario del sistema
type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	LastSeen time.Time `json:"last_seen"`
}

// Report representa un reporte de tráfico, accidente, etc.
type Report struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"` // "accident", "police", "traffic", "hazard"
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Description string    `json:"description"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	Votes       int       `json:"votes"`
}

// Route representa una ruta calculada
type Route struct {
	ID       int        `json:"id"`
	From     Location   `json:"from"`
	To       Location   `json:"to"`
	Points   []Location `json:"points"`
	Distance float64    `json:"distance"`
	Duration int        `json:"duration"` // en minutos
}

// Location representa una coordenada geográfica
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// TrafficData representa datos de tráfico en tiempo real
type TrafficData struct {
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	Speed      float64   `json:"speed"`
	Congestion string    `json:"congestion"` // "low", "medium", "high"
	Timestamp  time.Time `json:"timestamp"`
}

// NominatimResponse estructura para la respuesta de geocodificación
type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// WebSocketMessage mensaje para comunicación WebSocket
type WebSocketMessage struct {
	Type         string      `json:"type"`
	UsersOnline  int         `json:"users_online,omitempty"`
	TotalReports int         `json:"total_reports,omitempty"`
	TrafficPoints int        `json:"traffic_points,omitempty"`
	Report       *Report     `json:"report,omitempty"`
	Reports      []*Report   `json:"reports,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}