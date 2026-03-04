package main

import (
	"basket-cost/internal/database"
	"basket-cost/internal/enricher"
	"basket-cost/internal/handlers"
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
	"http://localhost:4173": true,
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func securityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		next(w, r)
	}
}

func chain(h http.HandlerFunc) http.HandlerFunc {
	return securityHeadersMiddleware(corsMiddleware(h))
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "basket-cost.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	s := store.New(db)
	imp := ticket.NewImporter(ticket.NewExtractor(), ticket.NewMercadonaParser(), s)
	enr := enricher.New(s)
	enr.Start(context.Background())
	h := handlers.New(s, imp, enr)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/products", chain(h.SearchHandler))
	mux.HandleFunc("/api/products/", chain(h.ProductHandler))
	mux.HandleFunc("/api/tickets", chain(h.TicketHandler))
	mux.HandleFunc("/api/analytics", chain(h.AnalyticsHandler))

	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("Basket Cost API server running on http://localhost%s\n", port)
	log.Fatal(srv.ListenAndServe())
}
