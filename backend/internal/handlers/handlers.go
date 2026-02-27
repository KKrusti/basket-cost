// Package handlers implements the HTTP handlers for the Basket Cost API.
package handlers

import (
	"basket-cost/internal/store"
	"encoding/json"
	"net/http"
)

// Handlers holds the shared dependencies injected at startup.
type Handlers struct {
	store store.Store
}

// New returns a Handlers instance wired to the given Store.
func New(s store.Store) *Handlers {
	return &Handlers{store: s}
}

// SearchHandler handles GET /api/products?q=<query>
// Returns a list of products matching the search query.
func (h *Handlers) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	results, err := h.store.SearchProducts(query)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ProductHandler handles GET /api/products/<id>
// Returns full product details including price history.
func (h *Handlers) ProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/products/123
	id := r.URL.Path[len("/api/products/"):]
	if id == "" {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	product, err := h.store.GetProductByID(id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
