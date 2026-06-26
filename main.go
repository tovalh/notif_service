package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := loadConfig()
	log.Printf("config loaded: addr %s, poll every %ds", cfg.WSAddr, cfg.PollSeconds)

	hub := newHub(cfg.WSSecret)
	log.Printf("vigilando %d base(s) de datos", len(cfg.DBDSNs))

	// Una goroutine (poller) por cada BD. Cada una lleva su propio cursor y
	// empuja al mismo hub, que es seguro para escrituras concurrentes.
	for i, dsn := range cfg.DBDSNs {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("BD %d: open: %v", i+1, err)
		}
		defer db.Close()

		// sql.Open does NOT connect yet; Ping forces a real connection.
		if err := db.Ping(); err != nil {
			log.Fatalf("BD %d: no responde: %v", i+1, err)
		}

		// Cursor seed: el id más alto ya existente, para no re-empujar lo viejo.
		var maxID int
		if err := db.QueryRow("SELECT COALESCE(MAX(idusu_notificacion), 0) FROM usu_notificacion").Scan(&maxID); err != nil {
			log.Fatalf("BD %d: seed cursor: %v", i+1, err)
		}
		log.Printf("BD %d conectada, cursor inicial = %d", i+1, maxID)

		poller := newPoller(db, hub, maxID, fmt.Sprintf("BD %d", i+1))
		go poller.run(time.Duration(cfg.PollSeconds) * time.Second)
	}

	// Health endpoint.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/ws", hub.handleWS)
	http.HandleFunc("/notify", hub.handleNotify)

	// Start the HTTP server (blocks forever).
	log.Printf("listening on http://localhost%s/health", cfg.WSAddr)
	if err := http.ListenAndServe(cfg.WSAddr, nil); err != nil {
		log.Fatalf("server died: %v", err)
	}
}
