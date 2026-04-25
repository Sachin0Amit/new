package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/Sachin0Amit/new/pkg/logger"
)

// EventType defines the classification of a telemetry signal.
type EventType string

const (
	EventLog      EventType = "LOG"
	EventTask     EventType = "TASK"
	EventFleet    EventType = "FLEET"
	EventSensory  EventType = "SENSORY"
	EventMetrics  EventType = "METRICS"
)

// Message encapsulates a structured telemetry signal for the Command Center.
type Message struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For local development: allow all origins
	},
}

// Hub manages the pool of active WebSocket connections for real-time telemetry.
type Hub struct {
	mu      sync.Mutex
	clients map[*Client]bool
	logger  logger.Logger
}

// Client represents a single administrative interface connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		logger:  logger.New(),
	}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket", logger.ErrorF(err))
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	go client.writePump()
}

func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal broadcast message", logger.ErrorF(err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.hub.mu.Lock()
		delete(c.hub.clients, c)
		c.hub.mu.Unlock()
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
