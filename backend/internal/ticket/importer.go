package ticket

import (
	"fmt"
	"io"
	"strings"

	"basket-cost/internal/models"
)

// TicketStore is the subset of store.Store required by the Importer.
// Using a narrow interface keeps the ticket package decoupled from the full
// store package and makes testing easier.
type TicketStore interface {
	UpsertPriceRecord(name string, record models.PriceRecord) error
}

// ImportResult summarises the outcome of a single ticket import.
type ImportResult struct {
	// InvoiceNumber is the receipt identifier.
	InvoiceNumber string
	// LinesImported is the number of product lines successfully persisted.
	LinesImported int
}

// Importer orchestrates PDF extraction → parsing → persistence.
type Importer struct {
	extractor PDFExtractor
	parser    Parser
	store     TicketStore
}

// NewImporter wires up the three collaborators.
func NewImporter(extractor PDFExtractor, parser Parser, store TicketStore) *Importer {
	return &Importer{
		extractor: extractor,
		parser:    parser,
		store:     store,
	}
}

// Import reads a PDF from r, parses it as a Mercadona receipt, and persists
// each product line as a price record.
// r must implement io.ReaderAt; use bytes.NewReader for in-memory data.
func (imp *Importer) Import(r io.ReaderAt, size int64) (*ImportResult, error) {
	text, err := imp.extractor.Extract(r, size)
	if err != nil {
		return nil, fmt.Errorf("extract pdf text: %w", err)
	}

	t, err := imp.parser.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("parse receipt: %w", err)
	}

	result := &ImportResult{InvoiceNumber: t.InvoiceNumber}

	var errs []string
	for _, line := range t.Lines {
		rec := models.PriceRecord{
			Date:  t.Date,
			Price: line.UnitPrice,
			Store: t.Store,
		}
		if err := imp.store.UpsertPriceRecord(line.Name, rec); err != nil {
			errs = append(errs, fmt.Sprintf("upsert %q: %v", line.Name, err))
			continue
		}
		result.LinesImported++
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("partial import errors: %s", strings.Join(errs, "; "))
	}
	return result, nil
}
