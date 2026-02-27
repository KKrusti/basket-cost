// Package handlers implements the HTTP handlers for the Basket Cost API.
package handlers

import (
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Handlers holds the shared dependencies injected at startup.
type Handlers struct {
	store    store.Store
	importer *ticket.Importer
}

// New returns a Handlers instance wired to the given Store and Importer.
func New(s store.Store, imp *ticket.Importer) *Handlers {
	return &Handlers{store: s, importer: imp}
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

// ticketResponse is the JSON body returned by TicketHandler on success.
type ticketResponse struct {
	InvoiceNumber string `json:"invoiceNumber"`
	LinesImported int    `json:"linesImported"`
}

// TicketHandler handles POST /api/tickets
// Accepts a multipart/form-data request with a "file" field containing a PDF.
// It parses the receipt and persists the extracted price data.
func (h *Handlers) TicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload to 10 MB to guard against oversized payloads.
	const maxUploadSize = 10 << 20
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Bad request: could not parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad request: missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error: could not read file", http.StatusInternalServerError)
		return
	}

	result, err := h.importer.Import(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		http.Error(w, "Unprocessable entity: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticketResponse{
		InvoiceNumber: result.InvoiceNumber,
		LinesImported: result.LinesImported,
	})
}
