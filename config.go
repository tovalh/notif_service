package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all the service settings.
type Config struct {
	DBDSNs      []string // una o más conexiones MySQL (una por BD a vigilar)
	WSAddr      string   // listen address, e.g. ":8081"
	WSSecret    string   // shared secret to validate tokens
	PollSeconds int      // how often to poll the DB
}

// loadConfig reads the env vars and builds a Config.
func loadConfig() Config {
	// Load .env if present; ignore the error (prod has no .env).
	_ = godotenv.Load()

	return Config{
		DBDSNs:      collectDSNs(),
		WSAddr:      envOr("WS_ADDR", ":8081"),
		WSSecret:    mustGet("WS_SECRET"),
		PollSeconds: envIntOr("POLL_SECONDS", 5),
	}
}

// collectDSNs junta los DSN definidos: DB_DSN (principal) y DB_DSN_2..DB_DSN_4.
// Los vacíos se ignoran, así puedes agregar BDs cuando las necesites sin tocar
// el código: basta con rellenar la variable de entorno. Exige al menos uno.
func collectDSNs() []string {
	keys := []string{"DB_DSN", "DB_DSN_2", "DB_DSN_3", "DB_DSN_4"}
	var out []string
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		log.Fatalf("config: no hay ninguna DB_DSN definida (revisa DB_DSN / DB_DSN_2..4)")
	}
	return out
}

// envOr returns the env var, or def if empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envIntOr returns the env var as int, or def if empty/invalid.
func envIntOr(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("config: %s=%q is not a number, using %d", key, v, def)
		return def
	}
	return n
}

// mustGet returns the env var, or exits if it is missing.
func mustGet(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("config: missing required env var %s", key)
	}
	return v
}
