package handlers

import (
	"encoding/json"
	"fmt"
	"gowaze/models"
	"gowaze/services"
	"gowaze/utils"
	"net/http"
	"strconv"
)

// APIHandler maneja las rutas de la API REST
type APIHandler struct {
	storage   *services.Storage
	wsService *services.WebSocketService
}

// NewAPIHandler crea una nueva instancia del handler de API
func NewAPIHandler(storage *services.Storage, wsService *services.WebSocketService) *APIHandler {
	return &APIHandler{
		storage:   storage,
		wsService: wsService,
	}
}

// CreateUserHandler maneja la creación/actualización de usuarios
func (h *APIHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lng, _ := strconv.ParseFloat(r.FormValue("lng"), 64)

	if username == "" {
		http.Error(w, "Username es requerido", http.StatusBadRequest)
		return
	}

	user := h.storage.CreateUser(username, lat, lng)

	// Broadcast actualización de estadísticas
	h.wsService.BroadcastStats()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div style="color: green; margin-top: 10px;">✅ Usuario "%s" registrado en (%.6f, %.6f)</div>`, 
		user.Username, user.Lat, user.Lng)
}

// CreateReportHandler maneja la creación de reportes
func (h *APIHandler) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	reportType := r.FormValue("type")
	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lng, _ := strconv.ParseFloat(r.FormValue("lng"), 64)
	description := r.FormValue("description")

	if reportType == "" {
		http.Error(w, "Tipo de reporte es requerido", http.StatusBadRequest)
		return
	}

	// Validar tipo de reporte
	validTypes := map[string]bool{
		"accident": true,
		"police":   true,
		"traffic":  true,
		"hazard":   true,
	}

	if !validTypes[reportType] {
		http.Error(w, "Tipo de reporte inválido", http.StatusBadRequest)
		return
	}

	report := h.storage.CreateReport(reportType, lat, lng, description, 1) // UserID por defecto

	// Broadcast nuevo reporte
	h.wsService.BroadcastNewReport(report)

	// Devolver lista actualizada de reportes
	h.GetReportsHandler(w, r)
}

// GetReportsHandler maneja la obtención de reportes
func (h *APIHandler) GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	reports := h.storage.GetRecentReports()

	w.Header().Set("Content-Type", "text/html")
	html := `<div class="reports-list">`

	if len(reports) == 0 {
		html += `<div style="text-align: center; color: #666; padding: 20px;">No hay reportes recientes</div>`
	}

	for _, report := range reports {
		icon := getReportIcon(report.Type)

		html += fmt.Sprintf(`
			<div class="report-item">
				<div class="report-type">%s %s</div>
				<div>%s</div>
				<div class="coordinates">📍 %.6f, %.6f</div>
				<div style="color: #666; font-size: 0.8em;">%s | 👍 %d votos</div>
			</div>
		`, icon, report.Type, report.Description, report.Lat, report.Lng, 
		   report.CreatedAt.Format("15:04"), report.Votes)
	}

	html += `</div>`
	fmt.Fprint(w, html)
}

// CalculateRouteHandler maneja el cálculo de rutas
func (h *APIHandler) CalculateRouteHandler(w http.ResponseWriter, r *http.Request) {
	fromLat, _ := strconv.ParseFloat(r.FormValue("from_lat"), 64)
	fromLng, _ := strconv.ParseFloat(r.FormValue("from_lng"), 64)
	toLat, _ := strconv.ParseFloat(r.FormValue("to_lat"), 64)
	toLng, _ := strconv.ParseFloat(r.FormValue("to_lng"), 64)

	// Validar coordenadas
	if fromLat == 0 || fromLng == 0 || toLat == 0 || toLng == 0 {
		http.Error(w, "Coordenadas inválidas", http.StatusBadRequest)
		return
	}

	// Calcular distancia usando fórmula haversine
	distance := utils.HaversineDistance(fromLat, fromLng, toLat, toLng)

	// Estimar duración (asumiendo velocidad promedio de 50 km/h en ciudad)
	duration := int(distance / 50 * 60) // en minutos

	// Simular puntos de ruta (línea recta dividida en segmentos)
	points := make([]models.Location, 0)
	segments := 10
	for i := 0; i <= segments; i++ {
		ratio := float64(i) / float64(segments)
		lat := fromLat + (toLat-fromLat)*ratio
		lng := fromLng + (toLng-fromLng)*ratio
		points = append(points, models.Location{Lat: lat, Lng: lng})
	}

	route := models.Route{
		From:     models.Location{Lat: fromLat, Lng: fromLng},
		To:       models.Location{Lat: toLat, Lng: toLng},
		Points:   points,
		Distance: distance,
		Duration: duration,
	}

	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
		<div class="route-info">
			<h4>📍 Ruta Calculada</h4>
			<p><strong>📏 Distancia:</strong> %.2f km</p>
			<p><strong>⏱️ Tiempo estimado:</strong> %d minutos</p>
			<p><strong>🅰️ Desde:</strong> %.6f, %.6f</p>
			<p><strong>🅱️ Hasta:</strong> %.6f, %.6f</p>
			<p><strong>📊 Puntos de ruta:</strong> %d</p>
			<div style="margin-top: 10px;">
				<small style="color: #666;">💡 Usando algoritmo Haversine + OpenStreetMap</small>
			</div>
		</div>
	`, route.Distance, route.Duration, fromLat, fromLng, toLat, toLng, len(route.Points))

	fmt.Fprint(w, html)
}

// GeocodeHandler maneja la geocodificación de direcciones
func (h *APIHandler) GeocodeHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Parámetro 'address' requerido", http.StatusBadRequest)
		return
	}

	// Llamar a la API de Nominatim
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?format=json&q=%s&limit=1", address)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Error llamando a la API de geocodificación", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var results []models.NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		http.Error(w, "Error procesando respuesta de geocodificación", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// getReportIcon retorna el emoji correspondiente al tipo de reporte
func getReportIcon(reportType string) string {
	icons := map[string]string{
		"accident": "🚗",
		"police":   "👮",
		"traffic":  "🚦",
		"hazard":   "⚠️",
	}
	
	if icon, exists := icons[reportType]; exists {
		return icon
	}
	return "📍" // Icono por defecto
}