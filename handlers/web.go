package handlers

import (
	"html/template"
	"net/http"
)

// WebHandler maneja las rutas del frontend
type WebHandler struct {
	template *template.Template
}

// NewWebHandler crea una nueva instancia del handler web
func NewWebHandler() *WebHandler {
	// Cargar template desde archivo
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		// Si no existe el archivo, usar template embebido
		tmpl = template.Must(template.New("index").Parse(embeddedTemplate))
	}

	return &WebHandler{
		template: tmpl,
	}
}

// HomeHandler maneja la página principal
func (h *WebHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	data := struct {
		Title string
	}{
		Title: "GoWaze - Navegación en Tiempo Real",
	}

	if err := h.template.Execute(w, data); err != nil {
		http.Error(w, "Error renderizando template", http.StatusInternalServerError)
		return
	}
}

// Template HTML embebido como fallback
const embeddedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    
    <!-- Leaflet (OpenStreetMap) -->
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
    
    <!-- Leaflet Routing Machine -->
    <link rel="stylesheet" href="https://unpkg.com/leaflet-routing-machine@3.2.12/dist/leaflet-routing-machine.css" />
    <script src="https://unpkg.com/leaflet-routing-machine@3.2.12/dist/leaflet-routing-machine.js"></script>
    
    <!-- Estilos CSS -->
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <div id="status" class="disconnected">Desconectado</div>
    
    <div class="container">
        <div class="header">
            <h1>🚗 GoWaze</h1>
            <p>Sistema de navegación con mapas reales en tiempo real</p>
        </div>

        <!-- Estadísticas -->
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number" id="users-online">0</div>
                <div>Usuarios Online</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="total-reports">0</div>
                <div>Reportes Activos</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="traffic-points">0</div>
                <div>Puntos de Tráfico</div>
            </div>
        </div>

        <div class="main-layout">
            <!-- Mapa Principal -->
            <div class="card">
                <h3>🗺️ Mapa Interactivo - OpenStreetMap</h3>
                <div class="map-container">
                    <div id="map"></div>
                </div>
                <div style="margin-top: 10px; font-size: 0.9em; color: #666;">
                    <strong>Controles:</strong> Click derecho para reportar • Arrastra para mover • Scroll para zoom
                </div>
            </div>

            <!-- Sidebar con controles -->
            <div class="sidebar">
                <!-- Ubicación del Usuario -->
                <div class="card">
                    <h3>👤 Tu Ubicación</h3>
                    <form hx-post="/api/users" hx-target="#user-status" hx-swap="innerHTML">
                        <div class="form-group">
                            <label for="username">Usuario:</label>
                            <input type="text" id="username" name="username" placeholder="Tu nombre" required>
                        </div>
                        <div class="compact-form">
                            <input type="number" id="lat" name="lat" step="0.000001" value="14.0818" placeholder="Latitud" required>
                            <button type="button" class="btn btn-secondary" onclick="getLocation()" style="width: auto; padding: 10px;">GPS</button>
                        </div>
                        <input type="number" id="lng" name="lng" step="0.000001" value="-87.2068" placeholder="Longitud" required style="margin-top: 10px;">
                        <button type="submit" class="btn">📍 Actualizar</button>
                    </form>
                    <div id="user-status"></div>
                </div>

                <!-- Búsqueda de Lugares -->
                <div class="card">
                    <h3>🔍 Buscar Lugar</h3>
                    <div class="form-group">
                        <input type="text" id="search-input" placeholder="Ej: San Pedro Sula, Honduras">
                        <button type="button" class="btn btn-secondary" onclick="searchPlace()">Buscar</button>
                    </div>
                    <div id="search-results"></div>
                </div>

                <!-- Crear Reporte -->
                <div class="card">
                    <h3>🚨 Nuevo Reporte</h3>
                    <form hx-post="/api/reports" hx-target="#reports-container" hx-swap="innerHTML">
                        <div class="form-group">
                            <select id="report-type" name="type" required>
                                <option value="accident">🚗 Accidente</option>
                                <option value="police">👮 Policía</option>
                                <option value="traffic">🚦 Tráfico</option>
                                <option value="hazard">⚠️ Peligro</option>
                            </select>
                        </div>
                        <input type="hidden" id="report-lat" name="lat" value="14.0818">
                        <input type="hidden" id="report-lng" name="lng" value="-87.2068">
                        <div class="form-group">
                            <textarea id="description" name="description" rows="2" placeholder="Describe lo que está pasando..."></textarea>
                        </div>
                        <button type="submit" class="btn">📢 Reportar</button>
                    </form>
                    <small style="color: #666;">Tip: Click derecho en el mapa para establecer ubicación</small>
                </div>

                <!-- Calcular Ruta -->
                <div class="card">
                    <h3>🧭 Ruta</h3>
                    <button type="button" class="btn btn-secondary" onclick="calculateRoute()">📍 Calcular Ruta</button>
                    <button type="button" class="btn btn-danger" onclick="clearRoute()">🗑️ Limpiar</button>
                    <div id="route-info"></div>
                </div>

                <!-- Reportes Recientes -->
                <div class="card">
                    <h3>📋 Reportes</h3>
                    <div id="reports-container" hx-get="/api/reports" hx-trigger="load, every 15s">
                        <div class="loading">Cargando reportes...</div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- JavaScript principal -->
    <script src="/static/js/app.js"></script>
</body>
</html>`