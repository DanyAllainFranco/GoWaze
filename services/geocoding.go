package services

import (
	"encoding/json"
	"fmt"
	"gowaze/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GeocodingService maneja las operaciones de geocodificación
type GeocodingService struct {
	client   *http.Client
	baseURL  string
	userAgent string
	rateLimit time.Duration
	lastRequest time.Time
}

// NewGeocodingService crea una nueva instancia del servicio de geocodificación
func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:   "https://nominatim.openstreetmap.org",
		userAgent: "GoWaze/1.0 (Navigation App)",
		rateLimit: time.Second, // Nominatim requiere máximo 1 request por segundo
	}
}

// GeocodingRequest estructura para peticiones de geocodificación
type GeocodingRequest struct {
	Query       string  `json:"query"`
	Limit       int     `json:"limit"`
	CountryCode string  `json:"countrycodes,omitempty"`
	Language    string  `json:"accept-language,omitempty"`
	Bounded     bool    `json:"bounded,omitempty"`
	ViewBox     *ViewBox `json:"viewbox,omitempty"`
}

// ViewBox define un área de búsqueda limitada
type ViewBox struct {
	MinLat float64 `json:"min_lat"`
	MinLng float64 `json:"min_lng"`
	MaxLat float64 `json:"max_lat"`
	MaxLng float64 `json:"max_lng"`
}

// GeocodingResult resultado de geocodificación enriquecido
type GeocodingResult struct {
	PlaceID     int     `json:"place_id"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	DisplayName string  `json:"display_name"`
	Type        string  `json:"type"`
	Class       string  `json:"class"`
	Importance  float64 `json:"importance"`
	Icon        string  `json:"icon,omitempty"`
	Address     Address `json:"address,omitempty"`
	BoundingBox []string `json:"boundingbox,omitempty"`
}

// Address estructura detallada de dirección
type Address struct {
	HouseNumber  string `json:"house_number,omitempty"`
	Road         string `json:"road,omitempty"`
	Neighbourhood string `json:"neighbourhood,omitempty"`
	Suburb       string `json:"suburb,omitempty"`
	City         string `json:"city,omitempty"`
	Municipality string `json:"municipality,omitempty"`
	County       string `json:"county,omitempty"`
	State        string `json:"state,omitempty"`
	Postcode     string `json:"postcode,omitempty"`
	Country      string `json:"country,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

// SearchPlaces busca lugares usando la API de Nominatim
func (gs *GeocodingService) SearchPlaces(req GeocodingRequest) ([]GeocodingResult, error) {
	// Respetar rate limit
	if err := gs.waitForRateLimit(); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Validar query
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query no puede estar vacío")
	}

	// Configurar valores por defecto
	if req.Limit == 0 {
		req.Limit = 5
	}
	if req.Limit > 50 {
		req.Limit = 50 // Límite máximo de Nominatim
	}

	// Construir URL
	searchURL := gs.buildSearchURL(req)

	// Realizar petición HTTP
	resp, err := gs.makeRequest(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error en petición HTTP: %w", err)
	}
	defer resp.Body.Close()

	// Validar respuesta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parsear respuesta JSON
	var rawResults []models.NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResults); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Convertir a resultado enriquecido
	results := make([]GeocodingResult, 0, len(rawResults))
	for _, raw := range rawResults {
		result, err := gs.convertToGeocodingResult(raw)
		if err != nil {
			continue // Saltar resultados inválidos
		}
		results = append(results, result)
	}

	return results, nil
}

// ReverseGeocode convierte coordenadas en dirección
func (gs *GeocodingService) ReverseGeocode(lat, lng float64) (*GeocodingResult, error) {
	// Validar coordenadas
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, fmt.Errorf("coordenadas inválidas: lat=%f, lng=%f", lat, lng)
	}

	// Respetar rate limit
	if err := gs.waitForRateLimit(); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Construir URL para reverse geocoding
	reverseURL := fmt.Sprintf("%s/reverse?format=json&lat=%f&lon=%f&addressdetails=1",
		gs.baseURL, lat, lng)

	// Realizar petición
	resp, err := gs.makeRequest(reverseURL)
	if err != nil {
		return nil, fmt.Errorf("error en petición HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parsear respuesta
	var rawResult models.NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Convertir resultado
	result, err := gs.convertToGeocodingResult(rawResult)
	if err != nil {
		return nil, fmt.Errorf("error converting result: %w", err)
	}

	return &result, nil
}

// SearchNearby busca lugares cerca de una ubicación
func (gs *GeocodingService) SearchNearby(lat, lng float64, query string, radius float64) ([]GeocodingResult, error) {
	// Calcular viewbox basado en el radio (aproximado)
	// 1 grado ≈ 111 km
	degreeRadius := radius / 111.0

	viewBox := &ViewBox{
		MinLat: lat - degreeRadius,
		MinLng: lng - degreeRadius,
		MaxLat: lat + degreeRadius,
		MaxLng: lng + degreeRadius,
	}

	req := GeocodingRequest{
		Query:   query,
		Limit:   10,
		Bounded: true,
		ViewBox: viewBox,
	}

	return gs.SearchPlaces(req)
}

// GetPlaceDetails obtiene detalles completos de un lugar por PlaceID
func (gs *GeocodingService) GetPlaceDetails(placeID int) (*GeocodingResult, error) {
	if err := gs.waitForRateLimit(); err != nil {
		return nil, err
	}

	detailsURL := fmt.Sprintf("%s/details?format=json&place_id=%d&addressdetails=1",
		gs.baseURL, placeID)

	resp, err := gs.makeRequest(detailsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResult models.NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResult); err != nil {
		return nil, err
	}

	result, err := gs.convertToGeocodingResult(rawResult)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// buildSearchURL construye la URL de búsqueda
func (gs *GeocodingService) buildSearchURL(req GeocodingRequest) string {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("q", req.Query)
	params.Set("limit", strconv.Itoa(req.Limit))
	params.Set("addressdetails", "1")
	params.Set("extratags", "1")
	params.Set("namedetails", "1")

	if req.CountryCode != "" {
		params.Set("countrycodes", req.CountryCode)
	}

	if req.Language != "" {
		params.Set("accept-language", req.Language)
	}

	if req.Bounded && req.ViewBox != nil {
		viewbox := fmt.Sprintf("%f,%f,%f,%f",
			req.ViewBox.MinLng, req.ViewBox.MinLat,
			req.ViewBox.MaxLng, req.ViewBox.MaxLat)
		params.Set("viewbox", viewbox)
		params.Set("bounded", "1")
	}

	return fmt.Sprintf("%s/search?%s", gs.baseURL, params.Encode())
}

// makeRequest realiza una petición HTTP con headers apropiados
func (gs *GeocodingService) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Headers requeridos por Nominatim
	req.Header.Set("User-Agent", gs.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "es,en")

	gs.lastRequest = time.Now()
	return gs.client.Do(req)
}

// waitForRateLimit espera el tiempo necesario para respetar el rate limit
func (gs *GeocodingService) waitForRateLimit() error {
	elapsed := time.Since(gs.lastRequest)
	if elapsed < gs.rateLimit {
		waitTime := gs.rateLimit - elapsed
		time.Sleep(waitTime)
	}
	return nil
}

// convertToGeocodingResult convierte NominatimResponse a GeocodingResult
func (gs *GeocodingService) convertToGeocodingResult(raw models.NominatimResponse) (GeocodingResult, error) {
	lat, err := strconv.ParseFloat(raw.Lat, 64)
	if err != nil {
		return GeocodingResult{}, fmt.Errorf("invalid latitude: %s", raw.Lat)
	}

	lng, err := strconv.ParseFloat(raw.Lon, 64)
	if err != nil {
		return GeocodingResult{}, fmt.Errorf("invalid longitude: %s", raw.Lon)
	}

	result := GeocodingResult{
		Lat:         lat,
		Lng:         lng,
		DisplayName: raw.DisplayName,
	}

	return result, nil
}

// ValidateCoordinates valida que las coordenadas sean válidas
func (gs *GeocodingService) ValidateCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// GetSuggestions obtiene sugerencias de autocompletado
func (gs *GeocodingService) GetSuggestions(query string, limit int) ([]string, error) {
	if limit == 0 {
		limit = 5
	}

	req := GeocodingRequest{
		Query: query,
		Limit: limit,
	}

	results, err := gs.SearchPlaces(req)
	if err != nil {
		return nil, err
	}

	suggestions := make([]string, 0, len(results))
	for _, result := range results {
		suggestions = append(suggestions, result.DisplayName)
	}

	return suggestions, nil
}

// SearchByCategory busca lugares por categoría
func (gs *GeocodingService) SearchByCategory(category string, lat, lng float64, radius float64) ([]GeocodingResult, error) {
	queries := map[string]string{
		"hospital":    "hospital",
		"pharmacy":    "farmacia",
		"gas_station": "gasolinera",
		"restaurant":  "restaurante",
		"bank":        "banco",
		"school":      "escuela",
		"police":      "policía",
		"fire_station": "bomberos",
		"hotel":       "hotel",
		"shopping":    "centro comercial",
	}

	query, exists := queries[category]
	if !exists {
		query = category
	}

	return gs.SearchNearby(lat, lng, query, radius)
}

// GetBoundingBox calcula el bounding box para una lista de coordenadas
func (gs *GeocodingService) GetBoundingBox(coordinates []models.Location) *ViewBox {
	if len(coordinates) == 0 {
		return nil
	}

	minLat, maxLat := coordinates[0].Lat, coordinates[0].Lat
	minLng, maxLng := coordinates[0].Lng, coordinates[0].Lng

	for _, coord := range coordinates[1:] {
		if coord.Lat < minLat {
			minLat = coord.Lat
		}
		if coord.Lat > maxLat {
			maxLat = coord.Lat
		}
		if coord.Lng < minLng {
			minLng = coord.Lng
		}
		if coord.Lng > maxLng {
			maxLng = coord.Lng
		}
	}

	return &ViewBox{
		MinLat: minLat,
		MinLng: minLng,
		MaxLat: maxLat,
		MaxLng: maxLng,
	}
}

// FormatAddress formatea una dirección de manera legible
func (gs *GeocodingService) FormatAddress(addr Address) string {
	parts := []string{}

	if addr.HouseNumber != "" && addr.Road != "" {
		parts = append(parts, addr.Road+" "+addr.HouseNumber)
	} else if addr.Road != "" {
		parts = append(parts, addr.Road)
	}

	if addr.Neighbourhood != "" {
		parts = append(parts, addr.Neighbourhood)
	}

	if addr.City != "" {
		parts = append(parts, addr.City)
	} else if addr.Municipality != "" {
		parts = append(parts, addr.Municipality)
	}

	if addr.State != "" {
		parts = append(parts, addr.State)
	}

	if addr.Country != "" {
		parts = append(parts, addr.Country)
	}

	return strings.Join(parts, ", ")
}

// GetPlaceIcon retorna el emoji apropiado para el tipo de lugar
func (gs *GeocodingService) GetPlaceIcon(placeType, class string) string {
	icons := map[string]string{
		// Tipos de lugares
		"hospital":     "🏥",
		"pharmacy":     "💊",
		"restaurant":   "🍽️",
		"cafe":         "☕",
		"bank":         "🏦",
		"atm":          "💳",
		"school":       "🏫",
		"university":   "🎓",
		"library":      "📚",
		"police":       "👮",
		"fire_station": "🚒",
		"hotel":        "🏨",
		"gas_station":  "⛽",
		"parking":      "🅿️",
		"shopping":     "🛒",
		"airport":      "✈️",
		"train_station": "🚆",
		"bus_station":  "🚌",
		"church":       "⛪",
		"mosque":       "🕌",
		"park":         "🌳",
		"beach":        "🏖️",
		"mountain":     "⛰️",
		"river":        "🌊",
		
		// Clases generales
		"amenity":     "📍",
		"shop":        "🏪",
		"tourism":     "🗺️",
		"leisure":     "🎯",
		"natural":     "🌿",
		"historic":    "🏛️",
		"transport":   "🚌",
	}

	// Buscar por tipo específico primero
	if icon, exists := icons[placeType]; exists {
		return icon
	}

	// Luego por clase general
	if icon, exists := icons[class]; exists {
		return icon
	}

	// Icono por defecto
	return "📍"
}

// CacheResult estructura para cache de resultados
type CacheResult struct {
	Results   []GeocodingResult
	Timestamp time.Time
	TTL       time.Duration
}

// SimpleCache cache en memoria simple
type SimpleCache struct {
	data map[string]CacheResult
}

// NewSimpleCache crea un nuevo cache simple
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data: make(map[string]CacheResult),
	}
}

// Get obtiene un resultado del cache
func (c *SimpleCache) Get(key string) ([]GeocodingResult, bool) {
	if result, exists := c.data[key]; exists {
		if time.Since(result.Timestamp) < result.TTL {
			return result.Results, true
		}
		// Cache expirado, eliminar
		delete(c.data, key)
	}
	return nil, false
}

// Set guarda un resultado en el cache
func (c *SimpleCache) Set(key string, results []GeocodingResult, ttl time.Duration) {
	c.data[key] = CacheResult{
		Results:   results,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// SearchWithCache busca lugares con cache
func (gs *GeocodingService) SearchWithCache(req GeocodingRequest, cache *SimpleCache) ([]GeocodingResult, error) {
	// Crear clave de cache
	cacheKey := fmt.Sprintf("search_%s_%d_%s", req.Query, req.Limit, req.CountryCode)
	
	// Intentar obtener del cache
	if cache != nil {
		if results, found := cache.Get(cacheKey); found {
			return results, nil
		}
	}

	// Si no está en cache, buscar
	results, err := gs.SearchPlaces(req)
	if err != nil {
		return nil, err
	}

	// Guardar en cache por 1 hora
	if cache != nil {
		cache.Set(cacheKey, results, time.Hour)
	}

	return results, nil
}

// Stats estadísticas del servicio de geocodificación
type GeocodingStats struct {
	TotalRequests    int           `json:"total_requests"`
	CacheHits        int           `json:"cache_hits"`
	CacheMisses      int           `json:"cache_misses"`
	AverageResponse  time.Duration `json:"average_response"`
	LastRequestTime  time.Time     `json:"last_request_time"`
	RateLimitWaits   int           `json:"rate_limit_waits"`
}

// GetStats retorna estadísticas del servicio
func (gs *GeocodingService) GetStats() GeocodingStats {
	return GeocodingStats{
		LastRequestTime: gs.lastRequest,
		// Otros stats se podrían implementar con contadores
	}
}