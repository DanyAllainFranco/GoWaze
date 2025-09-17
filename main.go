package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gowaze/handlers"
	"gowaze/services"

	"github.com/gorilla/mux"
)

func main() {
	// Inicializar servicios
	storage := services.NewStorage()
	trafficService := services.NewTrafficService(storage)
	wsService := services.NewWebSocketService(storage)

	// Inicializar handlers
	apiHandler := handlers.NewAPIHandler(storage, wsService)
	webHandler := handlers.NewWebHandler()
	wsHandler := handlers.NewWebSocketHandler(wsService)

	// Datos de ejemplo iniciales
	storage.InitSampleData()

	// Iniciar servicios en background
	go trafficService.Start()
	go wsService.HandleBroadcast()
	go storage.StartCleanup()

	// Configurar rutas
	r := mux.NewRouter()

	// Rutas estáticas
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
		http.FileServer(http.Dir("static/"))))

	// Rutas frontend
	r.HandleFunc("/", webHandler.HomeHandler).Methods("GET")

	// API Routes
	r.HandleFunc("/api/users", apiHandler.CreateUserHandler).Methods("POST")
	r.HandleFunc("/api/reports", apiHandler.CreateReportHandler).Methods("POST")
	r.HandleFunc("/api/reports", apiHandler.GetReportsHandler).Methods("GET")
	r.HandleFunc("/api/routes", apiHandler.CalculateRouteHandler).Methods("POST")
	r.HandleFunc("/api/geocode", apiHandler.GeocodeHandler).Methods("GET")

	// WebSocket
	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Configurar servidor con timeouts
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor
	fmt.Println("🚗 GoWaze con mapas reales iniciado!")
	fmt.Println("📍 URL: http://localhost:8080")
	fmt.Println("🗺️  Mapas: OpenStreetMap + Leaflet")
	fmt.Println("📡 WebSocket: ws://localhost:8080/ws")
	fmt.Println("🌍 API de geocodificación: Nominatim OSM")
	fmt.Println("🎯 Ubicación por defecto: San Pedro Sula, Honduras")
	fmt.Println("📊 Características:")
	fmt.Println("   • Mapas interactivos reales")
	fmt.Println("   • Cálculo de rutas con Leaflet Routing")
	fmt.Println("   • Búsqueda de lugares")
	fmt.Println("   • Reportes en tiempo real")
	fmt.Println("   • Geolocalización GPS")
	fmt.Println("   • Simulador de tráfico")

	log.Fatal(srv.ListenAndServe())
}