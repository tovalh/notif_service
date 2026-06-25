package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Poller struct {
	db     *sql.DB
	hub    *Hub
	cursor int
	label  string // identifica la BD en los logs, ej. "BD 1"
}

func newPoller(db *sql.DB, hub *Hub, startCursor int, label string) *Poller {
	return &Poller{db: db, hub: hub, cursor: startCursor, label: label}
}

// pollOnce queries new rows and pushes a ding-dong for each.
func (p *Poller) pollOnce() {
	log.Printf("[%s] revisando (cursor = %d)", p.label, p.cursor)

	rows, err := p.db.Query(
		`SELECT idusu_notificacion, idgen_empresa, idusu_usuario
                   FROM usu_notificacion
                  WHERE idusu_notificacion > ?
                    AND habilitado = 1
                  ORDER BY idusu_notificacion ASC`, p.cursor)
	if err != nil {
		log.Printf("poll query failed: %v", err)
		return
	}
	defer rows.Close()

	found := 0
	for rows.Next() {
		var id, empresa, usuario int
		if err := rows.Scan(&id, &empresa, &usuario); err != nil {
			log.Printf("scan failed: %v", err)
			continue
		}
		found++

		msg := []byte(fmt.Sprintf(`{"tipo":"notif","id":%d}`, id))
		delivered := p.hub.send(clientKey{empresa: empresa, usuario: usuario}, msg)
		log.Printf("[%s] notif nueva %d -> (empresa %d, usuario %d): entregada a %d conexión(es)", p.label, id, empresa, usuario, delivered)

		p.cursor = id // rows come ASC, so the last one is the new max
	}

	if found == 0 {
		log.Printf("[%s] sin notificaciones nuevas", p.label)
	} else {
		log.Printf("[%s] %d notificación(es) nueva(s), cursor ahora = %d", p.label, found, p.cursor)
	}
}

func (p *Poller) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		p.pollOnce()
	}
}
