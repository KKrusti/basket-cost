package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"basket-cost/internal/database"
	"basket-cost/internal/handlers"
	"basket-cost/internal/models"
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
)

// newHandlers creates a Handlers instance backed by an in-memory SQLite DB
// pre-seeded with a small set of deterministic products.
func newHandlers(t *testing.T) *handlers.Handlers {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s := store.New(db)

	seed := []models.Product{
		{
			ID:       "1",
			Name:     "LECHE ENTERA HACENDADO 1L",
			Category: "Lácteos",
			PriceHistory: []models.PriceRecord{
				{Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), Price: 0.79, Store: "Mercadona"},
				{Date: time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC), Price: 0.89, Store: "Mercadona"},
			},
		},
		{
			ID:       "2",
			Name:     "PAN BIMBO INTEGRAL",
			Category: "Panadería",
			PriceHistory: []models.PriceRecord{
				{Date: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC), Price: 1.89, Store: "Carrefour"},
				{Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Price: 2.15, Store: "Carrefour"},
			},
		},
	}
	for _, p := range seed {
		if err := s.InsertProduct(p); err != nil {
			t.Fatalf("seed product %s: %v", p.ID, err)
		}
	}

	return handlers.New(s, nil)
}

// --- SearchHandler ---

func TestSearchHandler_MethodNotAllowed(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	h.SearchHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestSearchHandler_EmptyQuery_ReturnsAll(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()
	h.SearchHandler(w, req)
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
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	h.SearchHandler(w, req)
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
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=xyznonexistent", nil)
	w := httptest.NewRecorder()
	h.SearchHandler(w, req)
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
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products?q=leche", nil)
	w := httptest.NewRecorder()
	h.SearchHandler(w, req)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}

// --- ProductHandler ---

func TestProductHandler_MethodNotAllowed(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/api/products/1", nil)
	w := httptest.NewRecorder()
	h.ProductHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestProductHandler_MissingID_ReturnsBadRequest(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products/", nil)
	w := httptest.NewRecorder()
	h.ProductHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestProductHandler_NotFound(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products/9999", nil)
	w := httptest.NewRecorder()
	h.ProductHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestProductHandler_ValidID_ReturnsProduct(t *testing.T) {
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	w := httptest.NewRecorder()
	h.ProductHandler(w, req)
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
	h := newHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	w := httptest.NewRecorder()
	h.ProductHandler(w, req)
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}

// --- TicketHandler fakes ---

// fakeExtractor and fakeParser are defined here so tests stay self-contained.

type fakeTicketExtractor struct {
	text string
	err  error
}

func (f *fakeTicketExtractor) Extract(_ io.ReaderAt, _ int64) (string, error) {
	return f.text, f.err
}

type fakeTicketParser struct {
	t   *ticket.Ticket
	err error
}

func (f *fakeTicketParser) Parse(_ string) (*ticket.Ticket, error) {
	return f.t, f.err
}

// stubTicketStore implements ticket.TicketStore; it satisfies store.Store via
// embedding a *store.SQLiteStore so the same Handlers struct can hold both.
// For TicketHandler tests we only need UpsertPriceRecord, so we use a plain DB.

// newHandlersWithImporter creates a Handlers instance wired with the given Importer.
func newHandlersWithImporter(t *testing.T, imp *ticket.Importer) *handlers.Handlers {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	s := store.New(db)
	return handlers.New(s, imp)
}

// buildMultipartRequest creates a multipart POST request with a "file" field
// containing the given data.
func buildMultipartRequest(t *testing.T, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "ticket.pdf")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/tickets", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

// sampleImportTicket returns a minimal Ticket used by fake parsers in handler tests.
func sampleImportTicket() *ticket.Ticket {
	return &ticket.Ticket{
		Store:         "Mercadona",
		Date:          time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
		InvoiceNumber: "4144-017-284404",
		Lines: []ticket.TicketLine{
			{Name: "LECHE ENTERA HACENDADO 1L", UnitPrice: 0.89, Quantity: 1},
		},
	}
}

// --- TicketHandler tests ---

func TestTicketHandler_MethodNotAllowed(t *testing.T) {
	imp := ticket.NewImporter(&fakeTicketExtractor{}, &fakeTicketParser{t: sampleImportTicket()}, store.New(mustOpenMemDB(t)))
	h := newHandlersWithImporter(t, imp)
	req := httptest.NewRequest(http.MethodGet, "/api/tickets", nil)
	w := httptest.NewRecorder()
	h.TicketHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestTicketHandler_MissingFileField_ReturnsBadRequest(t *testing.T) {
	imp := ticket.NewImporter(&fakeTicketExtractor{}, &fakeTicketParser{t: sampleImportTicket()}, store.New(mustOpenMemDB(t)))
	h := newHandlersWithImporter(t, imp)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/tickets", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	w := httptest.NewRecorder()
	h.TicketHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTicketHandler_ValidPDF_ReturnsCreated(t *testing.T) {
	db := mustOpenMemDB(t)
	imp := ticket.NewImporter(
		&fakeTicketExtractor{text: "raw text"},
		&fakeTicketParser{t: sampleImportTicket()},
		store.New(db),
	)
	h := newHandlersWithImporter(t, imp)

	req := buildMultipartRequest(t, []byte("%PDF-1.4 fake"))
	w := httptest.NewRecorder()
	h.TicketHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTicketHandler_ValidPDF_ResponseJSON(t *testing.T) {
	db := mustOpenMemDB(t)
	imp := ticket.NewImporter(
		&fakeTicketExtractor{text: "raw text"},
		&fakeTicketParser{t: sampleImportTicket()},
		store.New(db),
	)
	h := newHandlersWithImporter(t, imp)

	req := buildMultipartRequest(t, []byte("%PDF-1.4 fake"))
	w := httptest.NewRecorder()
	h.TicketHandler(w, req)

	var resp struct {
		InvoiceNumber string `json:"invoiceNumber"`
		LinesImported int    `json:"linesImported"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.InvoiceNumber != "4144-017-284404" {
		t.Errorf("invoiceNumber: want %q, got %q", "4144-017-284404", resp.InvoiceNumber)
	}
	if resp.LinesImported != 1 {
		t.Errorf("linesImported: want 1, got %d", resp.LinesImported)
	}
}

func TestTicketHandler_ImporterError_ReturnsUnprocessable(t *testing.T) {
	imp := ticket.NewImporter(
		&fakeTicketExtractor{err: errors.New("corrupt pdf")},
		&fakeTicketParser{},
		store.New(mustOpenMemDB(t)),
	)
	h := newHandlersWithImporter(t, imp)
	req := buildMultipartRequest(t, []byte("not a pdf"))
	w := httptest.NewRecorder()
	h.TicketHandler(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

// mustOpenMemDB is a test helper that opens an in-memory SQLite database.
func mustOpenMemDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open mem DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
