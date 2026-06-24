package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type notifyPayload struct {
	ID      int `json:"id"`
	Empresa int `json:"idgen_empresa"`
	Usuario int `json:"idusu_usuario"`
}

func (h *Hub) handleNotify(w http.ResponseWriter, r *http.Request) {
	var p notifyPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Printf("notify: json inválido: %v", err)
		http.Error(w, "Bad json", http.StatusBadRequest)
		return
	}

	log.Printf("notify: toque recibido de PHP -> notif %d para (empresa %d, usuario %d)", p.ID, p.Empresa, p.Usuario)

	msg := []byte(fmt.Sprintf(`{"tipo":"notif","id":%d}`, p.ID))
	delivered := h.send(clientKey{empresa: p.Empresa, usuario: p.Usuario}, msg)
	log.Printf("notify: notif %d entregada a %d conexión(es)", p.ID, delivered)

	w.WriteHeader(http.StatusOK)
}
