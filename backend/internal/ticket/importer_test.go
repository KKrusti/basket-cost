package ticket_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"basket-cost/internal/models"
	"basket-cost/internal/ticket"
)

// --- Fakes ---

// fakeExtractor implements PDFExtractor and returns a fixed text string.
type fakeExtractor struct {
	text string
	err  error
}

func (f *fakeExtractor) Extract(_ io.ReaderAt, _ int64) (string, error) {
	return f.text, f.err
}

// fakeParser implements Parser and returns a fixed Ticket.
type fakeParser struct {
	t   *ticket.Ticket
	err error
}

func (f *fakeParser) Parse(_ string) (*ticket.Ticket, error) {
	return f.t, f.err
}

// fakeStore implements TicketStore and records all calls.
type fakeStore struct {
	records []models.PriceRecord
	names   []string
	err     error
}

func (f *fakeStore) UpsertPriceRecord(name string, record models.PriceRecord) error {
	if f.err != nil {
		return f.err
	}
	f.names = append(f.names, name)
	f.records = append(f.records, record)
	return nil
}

// --- Helpers ---

func sampleTicket() *ticket.Ticket {
	return &ticket.Ticket{
		Store:         "Mercadona",
		Date:          time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
		InvoiceNumber: "4144-017-284404",
		Lines: []ticket.TicketLine{
			{Name: "LECHE ENTERA HACENDADO 1L", UnitPrice: 0.89, Quantity: 1},
			{Name: "YOGUR NATURAL", UnitPrice: 0.35, Quantity: 2},
		},
	}
}

// --- Tests ---

func TestImporter_Import_Success(t *testing.T) {
	store := &fakeStore{}
	imp := ticket.NewImporter(
		&fakeExtractor{text: "raw text"},
		&fakeParser{t: sampleTicket()},
		store,
	)

	result, err := imp.Import(bytes.NewReader([]byte{}), 0)
	if err != nil {
		t.Fatalf("Import returned unexpected error: %v", err)
	}
	if result.LinesImported != 2 {
		t.Errorf("LinesImported: want 2, got %d", result.LinesImported)
	}
	if result.InvoiceNumber != "4144-017-284404" {
		t.Errorf("InvoiceNumber: want %q, got %q", "4144-017-284404", result.InvoiceNumber)
	}
	if len(store.names) != 2 {
		t.Errorf("expected 2 UpsertPriceRecord calls, got %d", len(store.names))
	}
}

func TestImporter_Import_ExtractorError(t *testing.T) {
	imp := ticket.NewImporter(
		&fakeExtractor{err: errors.New("pdf corrupt")},
		&fakeParser{},
		&fakeStore{},
	)
	_, err := imp.Import(bytes.NewReader([]byte{}), 0)
	if err == nil {
		t.Error("expected error from extractor, got nil")
	}
}

func TestImporter_Import_ParserError(t *testing.T) {
	imp := ticket.NewImporter(
		&fakeExtractor{text: "some text"},
		&fakeParser{err: errors.New("unrecognised format")},
		&fakeStore{},
	)
	_, err := imp.Import(bytes.NewReader([]byte{}), 0)
	if err == nil {
		t.Error("expected error from parser, got nil")
	}
}

func TestImporter_Import_StoreError_PartialResult(t *testing.T) {
	imp := ticket.NewImporter(
		&fakeExtractor{text: "some text"},
		&fakeParser{t: sampleTicket()},
		&fakeStore{err: errors.New("db locked")},
	)
	result, err := imp.Import(bytes.NewReader([]byte{}), 0)
	// Partial error is returned but result is non-nil.
	if err == nil {
		t.Error("expected partial error, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil result even on partial error")
	}
	if result.LinesImported != 0 {
		t.Errorf("expected 0 lines imported on store failure, got %d", result.LinesImported)
	}
}

func TestImporter_Import_PriceRecordDate(t *testing.T) {
	store := &fakeStore{}
	imp := ticket.NewImporter(
		&fakeExtractor{text: "text"},
		&fakeParser{t: sampleTicket()},
		store,
	)
	_, err := imp.Import(bytes.NewReader([]byte{}), 0)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	want := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	for _, rec := range store.records {
		if !rec.Date.Equal(want) {
			t.Errorf("record Date: want %s, got %s", want, rec.Date)
		}
	}
}
