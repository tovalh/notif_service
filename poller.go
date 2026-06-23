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
}

func newPoller(db *sql.DB, hub *Hub, startCursor int) *Poller {
	return &Poller{db: db, hub: hub, cursor: startCursor}
}

// pollOnce queries new rows and pushes a ding-dong for each.
func (p *Poller) pollOnce() {
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

	for rows.Next() {
		var id, empresa, usuario int
		if err := rows.Scan(&id, &empresa, &usuario); err != nil {
			log.Printf("scan failed: %v", err)
			continue
		}

		msg := []byte(fmt.Sprintf(`{"tipo":"notif","id":%d}`, id))
		p.hub.send(clientKey{empresa: empresa, usuario: usuario}, msg)
		log.Printf("pushed notif %d to (empresa %d, usuario %d)", id, empresa, usuario)

		p.cursor = id // rows come ASC, so the last one is the new max
	}
}

func (p *Poller) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		p.pollOnce()
	}
}
