// Package handlers implements the HTTP handlers for the Basket Cost API.
package handlers

import (
	"basket-cost/internal/models"
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const pdfMagic = "%PDF-"

// EnrichScheduler is the subset of *enricher.Enricher used by Handlers.
// Defined as an interface so tests can inject a fake without network calls.
type EnrichScheduler interface {
	Schedule()
}

type Handlers struct {
	store         store.Store
	importer      *ticket.Importer
	enricher      EnrichScheduler
	ticketLimiter *rate.Limiter
}

// New returns a Handlers instance. enr may be nil to skip post-import enrichment.
func New(s store.Store, imp *ticket.Importer, enr EnrichScheduler) *Handlers {
	return &Handlers{
		store:         s,
		importer:      imp,
		enricher:      enr,
		ticketLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 10),
	}
}

// NewWithLimiter creates a Handlers instance with a custom rate limiter.
// Intended for testing; production code should use New.
func NewWithLimiter(s store.Store, imp *ticket.Importer, enr EnrichScheduler, lim *rate.Limiter) *Handlers {
	return &Handlers{store: s, importer: imp, enricher: enr, ticketLimiter: lim}
}

func (h *Handlers) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	results, err := h.store.SearchProducts(r.URL.Query().Get("q"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("handlers: encode search response: %v", err)
	}
}

func (h *Handlers) ProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

type ticketResponse struct {
	InvoiceNumber string `json:"invoiceNumber"`
	LinesImported int    `json:"linesImported"`
}

type analyticsResponse struct {
	MostPurchased    []models.MostPurchasedProduct `json:"mostPurchased"`
	BiggestIncreases []models.PriceIncreaseProduct `json:"biggestIncreases"`
}

func (h *Handlers) TicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.ticketLimiter.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

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

	// Validate PDF magic bytes before invoking the parser to reject non-PDF uploads early.
	if len(data) < len(pdfMagic) || string(data[:len(pdfMagic)]) != pdfMagic {
		http.Error(w, "Unprocessable entity: file does not appear to be a valid PDF", http.StatusUnprocessableEntity)
		return
	}

	result, err := h.importer.Import(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Printf("handlers: ticket import failed for %q: %v", filename, err)
		http.Error(w, "Unprocessable entity: could not parse the PDF as a Mercadona receipt", http.StatusUnprocessableEntity)
		return
	}

	if err := h.store.MarkFileProcessed(filename, time.Now()); err != nil {
		// Non-fatal: the import succeeded; log and continue.
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

	// Concurrent Schedule calls are coalesced by the enricher, so batch uploads
	// trigger only one enrichment run.
	if h.enricher != nil {
		h.enricher.Schedule()
	}
}

const analyticsLimit = 10

func (h *Handlers) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mostPurchased, err := h.store.GetMostPurchased(analyticsLimit)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	biggestIncreases, err := h.store.GetBiggestPriceIncreases(analyticsLimit)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analyticsResponse{
		MostPurchased:    mostPurchased,
		BiggestIncreases: biggestIncreases,
	}); err != nil {
		log.Printf("handlers: encode analytics response: %v", err)
	}
}
