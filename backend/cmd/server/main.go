package main

import (
	"basket-cost/internal/database"
	"basket-cost/internal/handlers"
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"fmt"
	"log"
	"net/http"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Open (or create) the SQLite database file.
	db, err := database.Open("basket-cost.db")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	s := store.New(db)
	imp := ticket.NewImporter(ticket.NewExtractor(), ticket.NewMercadonaParser(), s)
	h := handlers.New(s, imp)
	mux := http.NewServeMux()

	// GET /api/products?q=<query> — search products
	mux.HandleFunc("/api/products", corsMiddleware(h.SearchHandler))

	// GET /api/products/<id> — get product detail with price history
	mux.HandleFunc("/api/products/", corsMiddleware(h.ProductHandler))

	// POST /api/tickets — upload a Mercadona PDF receipt
	mux.HandleFunc("/api/tickets", corsMiddleware(h.TicketHandler))

	port := ":8080"
	fmt.Printf("Basket Cost API server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
