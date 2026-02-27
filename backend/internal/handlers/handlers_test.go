package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"basket-cost/internal/handlers"
	"basket-cost/internal/models"
)

// --- SearchHandler ---

func TestSearchHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	handlers.SearchHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestSearchHandler_EmptyQuery_ReturnsAll(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()
	handlers.SearchHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var results []models.SearchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result for empty query")
	}
}

func TestSearchHandler_WithQuery_ReturnsMatches(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	handlers.SearchHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var results []models.SearchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'leche'")
	}
}

func TestSearchHandler_NoMatch_ReturnsEmptyArray(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=xyznonexistent", nil)
	w := httptest.NewRecorder()
	handlers.SearchHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var results []models.SearchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if results == nil {
		t.Error("expected empty array, not null")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchHandler_ContentTypeJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	handlers.SearchHandler(w, req)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}

// --- ProductHandler ---

func TestProductHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/products/1", nil)
	w := httptest.NewRecorder()
	handlers.ProductHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestProductHandler_MissingID_ReturnsBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/", nil)
	w := httptest.NewRecorder()
	handlers.ProductHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestProductHandler_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/9999", nil)
	w := httptest.NewRecorder()
	handlers.ProductHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestProductHandler_ValidID_ReturnsProduct(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	w := httptest.NewRecorder()
	handlers.ProductHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var product models.Product
	if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if product.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", product.ID)
	}
	if len(product.PriceHistory) == 0 {
		t.Error("product should have price history")
	}
}

func TestProductHandler_ContentTypeJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	w := httptest.NewRecorder()
	handlers.ProductHandler(w, req)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}
