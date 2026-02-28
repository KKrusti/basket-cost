// Package handlers implements the HTTP handlers for the Basket Cost API.
package handlers

import (
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// EnrichScheduler is the subset of *enricher.Enricher consumed by Handlers.
// Declaring it as an interface allows tests to inject a fake without network calls.
// Schedule signals the background enrichment worker to run once; concurrent
// calls are coalesced so the Mercadona API is not spammed.
type EnrichScheduler interface {
	Schedule()
}

// Handlers holds the shared dependencies injected at startup.
type Handlers struct {
	store    store.Store
	importer *ticket.Importer
	enricher EnrichScheduler
}

// New returns a Handlers instance wired to the given Store, Importer and EnrichScheduler.
// enr may be nil, in which case automatic enrichment after ticket import is skipped.
func New(s store.Store, imp *ticket.Importer, enr EnrichScheduler) *Handlers {
	return &Handlers{store: s, importer: imp, enricher: enr}
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
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("handlers: encode search response: %v", err)
	}
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
	if err := json.NewEncoder(w).Encode(product); err != nil {
		log.Printf("handlers: encode product response: %v", err)
	}
}

// ticketResponse is the JSON body returned by TicketHandler on success.
type ticketResponse struct {
	InvoiceNumber string `json:"invoiceNumber"`
	LinesImported int    `json:"linesImported"`
}

// TicketHandler handles POST /api/tickets
// Accepts a multipart/form-data request with a "file" field containing a PDF.
// It parses the receipt and persists the extracted price data.
// On success it triggers a background enrichment pass to update product image URLs.
// If the filename has already been imported a 409 Conflict is returned.
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

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad request: missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename

	// Reject files that have already been imported.
	already, err := h.store.IsFileProcessed(filename)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if already {
		http.Error(w, "Conflict: file already imported", http.StatusConflict)
		return
	}

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

	// Record the file as processed so future uploads of the same name are rejected.
	if err := h.store.MarkFileProcessed(filename, time.Now()); err != nil {
		// Non-fatal: log and continue — the import succeeded.
		log.Printf("handlers: could not mark file processed %q: %v", filename, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ticketResponse{
		InvoiceNumber: result.InvoiceNumber,
		LinesImported: result.LinesImported,
	}); err != nil {
		log.Printf("handlers: encode ticket response: %v", err)
	}

	// Signal the background enricher worker. Concurrent signals are coalesced
	// so a batch of ticket uploads triggers only one enrichment run.
	if h.enricher != nil {
		h.enricher.Schedule()
	}
}
