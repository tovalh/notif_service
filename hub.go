package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type clientKey struct {
	empresa int
	usuario int
}

type Hub struct {
	mu      sync.Mutex
	clients map[clientKey][]*websocket.Conn
	secret  string // shared secret used to validate connection tokens
}

func newHub(secret string) *Hub {
	return &Hub{
		clients: make(map[clientKey][]*websocket.Conn),
		secret:  secret,
	}
}

func (h *Hub) register(key clientKey, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[key] = append(h.clients[key], conn)
}

func (h *Hub) unregister(key clientKey, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var kept []*websocket.Conn
	for _, c := range h.clients[key] {
		if c != conn {
			kept = append(kept, c)
		}
	}

	if len(kept) == 0 {
		delete(h.clients, key)
	} else {
		h.clients[key] = kept
	}
}

func (h *Hub) send(key clientKey, msg []byte) {
	h.mu.Lock()
	conns := h.clients[key]
	h.mu.Unlock()

	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Failed to send message to client: %v", err)
		}
	}
}
