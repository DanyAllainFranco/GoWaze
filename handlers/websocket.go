package handlers

import (
	"gowaze/services"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketHandler maneja las conexiones WebSocket
type WebSocketHandler struct {
	wsService *services.WebSocketService
}

// NewWebSocketHandler crea una nueva instancia del handler WebSocket
func NewWebSocketHandler(wsService *services.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// HandleWebSocket maneja las conexiones WebSocket
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := h.wsService.GetUpgrader()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error actualizando conexión a WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Agregar cliente
	h.wsService.AddClient(conn)
	defer h.wsService.RemoveClient(conn)

	// Escuchar mensajes del cliente
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error WebSocket: %v", err)
			}
			break
		}

		// Procesar mensajes del cliente si es necesario
		h.handleClientMessage(msg)
	}
}

// handleClientMessage procesa mensajes recibidos del cliente
func (h *WebSocketHandler) handleClientMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Responder a ping para mantener conexión viva
		log.Printf("📡 Ping recibido de cliente WebSocket")
	case "request_stats":
		// Cliente solicita estadísticas actualizadas
		h.wsService.BroadcastStats()
	default:
		log.Printf("📨 Mensaje WebSocket desconocido: %s", msgType)
	}
}
