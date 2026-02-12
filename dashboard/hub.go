package dashboard

import (
	log "github.com/sirupsen/logrus"
)

// Hub maintains the set of active WebSocket clients and broadcasts messages to them.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the ticker to broadcast to clients.
	broadcast chan []byte

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's main loop, handling register/unregister/broadcast events.
// It is intended to be run as a goroutine. Panic recovery ensures a dashboard
// failure does not crash the trading bot (WS-05 isolation).
func (h *Hub) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Hub panic recovered: %v", r)
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Infof("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Infof("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
