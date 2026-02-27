package handlers

import (
	"basket-cost/internal/mockdata"
	"basket-cost/internal/models"
	"encoding/json"
	"net/http"
)

// SearchHandler handles GET /api/products?q=<query>
// Returns a list of products matching the search query.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	results := mockdata.SearchProducts(query)

	// Return empty array instead of null
	if results == nil {
		results = []models.SearchResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ProductHandler handles GET /api/products/<id>
// Returns full product details including price history.
func ProductHandler(w http.ResponseWriter, r *http.Request) {
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

	product := mockdata.GetProductByID(id)
	if product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
