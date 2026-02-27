package main

import (
	"basket-cost/internal/handlers"
	"fmt"
	"log"
	"net/http"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	mux := http.NewServeMux()

	// GET /api/products?q=<query> — search products
	mux.HandleFunc("/api/products", corsMiddleware(handlers.SearchHandler))

	// GET /api/products/<id> — get product detail with price history
	mux.HandleFunc("/api/products/", corsMiddleware(handlers.ProductHandler))

	port := ":8080"
	fmt.Printf("Basket Cost API server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
