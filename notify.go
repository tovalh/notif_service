package main

import (
	"encoding/json"
	"fmt"
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
		http.Error(w, "Bad json", http.StatusBadRequest)
		return
	}

	msg := []byte(fmt.Sprintf(`{"tipo":"notif","id":%d}`, p.ID))
	h.send(clientKey{empresa: p.Empresa, usuario: p.Usuario}, msg)

	w.WriteHeader(http.StatusOK)
}
