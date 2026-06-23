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

	db, err := sql.Open("mysql", cfg.DBDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// sql.Open does NOT connect yet; Ping forces a real connection.
	if err := db.Ping(); err != nil {
		log.Fatalf("cannot reach DB: %v", err)
	}
	log.Println("DB connected!")

	// First real read: the cursor seed (highest existing id).
	var maxID int
	err = db.QueryRow("SELECT COALESCE(MAX(idusu_notificacion), 0) FROM usu_notificacion").Scan(&maxID)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	log.Printf("max notification id = %d", maxID)

	// Health endpoint.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	hub := newHub(cfg.WSSecret)

	poller := newPoller(db, hub, maxID)
	go poller.run(time.Duration(cfg.PollSeconds) * time.Second)

	http.HandleFunc("/ws", hub.handleWS)
	http.HandleFunc("/notify", hub.handleNotify)

	// Start the HTTP server (blocks forever).
	log.Printf("listening on http://localhost%s/health", cfg.WSAddr)
	if err := http.ListenAndServe(cfg.WSAddr, nil); err != nil {
		log.Fatalf("server died: %v", err)
	}
}
