package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
	// Validate the token BEFORE upgrading: once the connection becomes a
	// WebSocket we can no longer reply with a normal HTTP error code.
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "falta token", http.StatusUnauthorized)
		return
	}

	usuario, empresa, err := ParseToken(token, h.secret)
	if err != nil {
		log.Printf("token rechazado: %v", err)
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	key := clientKey{empresa: empresa, usuario: usuario}
	log.Printf("cliente conectado (empresa %d, usuario %d)", empresa, usuario)
	h.register(key, conn)
	defer h.unregister(key, conn)

	// Block reading until the client goes away. We don't expect messages
	// from the browser; this loop just detects when the connection closes.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("cliente desconectado (empresa %d, usuario %d)", empresa, usuario)
			return
		}
	}
}