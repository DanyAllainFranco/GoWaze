package services

import (
	"encoding/json"
	"gowaze/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketService maneja las conexiones WebSocket
type WebSocketService struct {
	storage   *Storage
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	upgrader  websocket.Upgrader
}

// NewWebSocketService crea una nueva instancia del servicio WebSocket
func NewWebSocketService(storage *Storage) *WebSocketService {
	return &WebSocketService{
		storage:   storage,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // En producción, implementar verificación de origen
			},
		},
	}
}

// HandleBroadcast maneja el envío de mensajes broadcast
func (ws *WebSocketService) HandleBroadcast() {
	for {
		msg := <-ws.broadcast
		for client := range ws.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("Error enviando mensaje WebSocket: %v", err)
				client.Close()
				delete(ws.clients, client)
			}
		}
	}
}

// AddClient agrega un nuevo cliente WebSocket
func (ws *WebSocketService) AddClient(conn *websocket.Conn) {
	ws.clients[conn] = true

	// Enviar estadísticas iniciales
	ws.SendStatsToClient(conn)

	log.Printf("🔌 Nuevo cliente WebSocket conectado. Total: %d", len(ws.clients))
}

// RemoveClient remueve un cliente WebSocket
func (ws *WebSocketService) RemoveClient(conn *websocket.Conn) {
	if _, ok := ws.clients[conn]; ok {
		delete(ws.clients, conn)
		conn.Close()
		log.Printf("🔌 Cliente WebSocket desconectado. Total: %d", len(ws.clients))
	}
}

// BroadcastStats envía estadísticas a todos los clientes
func (ws *WebSocketService) BroadcastStats() {
	usersOnline, totalReports, trafficPoints := ws.storage.GetStats()

	msg := models.WebSocketMessage{
		Type:          "stats",
		UsersOnline:   usersOnline,
		TotalReports:  totalReports,
		TrafficPoints: trafficPoints,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error serializando estadísticas: %v", err)
		return
	}

	ws.broadcast <- data
}

// BroadcastNewReport envía un nuevo reporte a todos los clientes
func (ws *WebSocketService) BroadcastNewReport(report *models.Report) {
	reports := ws.storage.GetRecentReports()

	msg := models.WebSocketMessage{
		Type:    "new_report",
		Report:  report,
		Reports: reports,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error serializando reporte: %v", err)
		return
	}

	ws.broadcast <- data

	// También enviar estadísticas actualizadas
	ws.BroadcastStats()
}

// SendStatsToClient envía estadísticas a un cliente específico
func (ws *WebSocketService) SendStatsToClient(conn *websocket.Conn) {
	usersOnline, totalReports, trafficPoints := ws.storage.GetStats()
	reports := ws.storage.GetRecentReports()

	msg := models.WebSocketMessage{
		Type:          "stats",
		UsersOnline:   usersOnline,
		TotalReports:  totalReports,
		TrafficPoints: trafficPoints,
		Reports:       reports,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error enviando estadísticas iniciales: %v", err)
	}
}

// GetUpgrader retorna el upgrader de WebSocket
func (ws *WebSocketService) GetUpgrader() *websocket.Upgrader {
	return &ws.upgrader
}

// GetClientCount retorna el número de clientes conectados
func (ws *WebSocketService) GetClientCount() int {
	return len(ws.clients)
}
