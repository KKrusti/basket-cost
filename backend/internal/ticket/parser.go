package ticket

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Parser is the contract for turning raw receipt text into a structured Ticket.
type Parser interface {
	// Parse converts the plain-text content of a receipt into a Ticket.
	// It returns an error if the text does not match the expected format.
	Parse(text string) (*Ticket, error)
}

// MercadonaParser parses receipts from Mercadona (Caldes de Montbui branch).
// Receipts are written in Catalan.
type MercadonaParser struct{}

// NewMercadonaParser returns a ready-to-use MercadonaParser.
func NewMercadonaParser() *MercadonaParser {
	return &MercadonaParser{}
}

// Compiled regexes reused across calls.
var (
	// Header: date line — "09/02/2026 19:43   OP: 2570140"
	reDateLine = regexp.MustCompile(`(\d{2}/\d{2}/\d{4})`)

	// Header: invoice line — "FACTURA SIMPLIFICADA: 4144-017-284404"
	reInvoice = regexp.MustCompile(`FACTURA SIMPLIFICADA:\s*(\S+)`)

	// Product by unit, qty=1 (no total column):
	//   "1   LECHE ENTERA HACENDADO 1L   0,89"
	reUnitSingle = regexp.MustCompile(`^1\s{2,}(.+?)\s{2,}(\d+,\d{2})\s*$`)

	// Product by unit, qty>1 (has total column):
	//   "3   AGUA MINERAL 1,5L   0,45   1,35"
	reUnitMulti = regexp.MustCompile(`^(\d+)\s{2,}(.+?)\s{2,}(\d+,\d{2})\s{2,}\d+,\d{2}\s*$`)

	// Product by weight — second continuation line:
	//   "0,354 kg   6,99 €/kg   2,47"
	reWeightLine = regexp.MustCompile(`^(\d+,\d+)\s*kg\s+(\d+,\d{2})\s*€/kg\s+\d+,\d{2}\s*$`)

	// Footer sentinel — everything from here on is ignored.
	reFooter = regexp.MustCompile(`TOTAL\s*\(€\)`)
)

// Parse implements Parser for Mercadona receipts.
func (p *MercadonaParser) Parse(text string) (*Ticket, error) {
	lines := splitLines(text)

	t := &Ticket{Store: "Mercadona"}

	// --- Extract header fields ---
	for _, line := range lines {
		if t.Date.IsZero() {
			if m := reDateLine.FindStringSubmatch(line); m != nil {
				d, err := time.Parse("02/01/2006", m[1])
				if err == nil {
					t.Date = d
				}
			}
		}
		if t.InvoiceNumber == "" {
			if m := reInvoice.FindStringSubmatch(line); m != nil {
				t.InvoiceNumber = m[1]
			}
		}
		if !t.Date.IsZero() && t.InvoiceNumber != "" {
			break
		}
	}

	if t.Date.IsZero() {
		return nil, fmt.Errorf("could not find date in receipt")
	}

	// --- Extract product lines ---
	// We need to detect the body section: it starts after the column-header line
	// "Descripció   P. Unit   Import" and ends at the TOTAL sentinel.
	inBody := false
	pendingWeightProduct := "" // product name waiting for a weight continuation line

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect footer — stop processing products.
		if reFooter.MatchString(trimmed) {
			break
		}

		// Detect start of body section.
		if strings.Contains(trimmed, "Descripció") && strings.Contains(trimmed, "P. Unit") {
			inBody = true
			continue
		}
		if !inBody {
			continue
		}
		if trimmed == "" {
			continue
		}

		// --- Weight continuation line ---
		// If the previous product was by weight, this line carries price/kg.
		if pendingWeightProduct != "" {
			if m := reWeightLine.FindStringSubmatch(trimmed); m != nil {
				pricePerKg, err := parsePrice(m[2])
				if err == nil {
					t.Lines = append(t.Lines, TicketLine{
						Name:      pendingWeightProduct,
						UnitPrice: pricePerKg,
						Quantity:  1,
					})
				}
				pendingWeightProduct = ""
				continue
			}
			// Not a weight line; discard pending (unexpected format).
			pendingWeightProduct = ""
		}

		// --- Unit product, qty > 1 ---
		if m := reUnitMulti.FindStringSubmatch(trimmed); m != nil {
			qty, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			price, err := parsePrice(m[3])
			if err != nil {
				continue
			}
			t.Lines = append(t.Lines, TicketLine{
				Name:      strings.TrimSpace(m[2]),
				UnitPrice: price,
				Quantity:  qty,
			})
			continue
		}

		// --- Unit product, qty = 1 ---
		if m := reUnitSingle.FindStringSubmatch(trimmed); m != nil {
			price, err := parsePrice(m[2])
			if err != nil {
				continue
			}
			t.Lines = append(t.Lines, TicketLine{
				Name:      strings.TrimSpace(m[1]),
				UnitPrice: price,
				Quantity:  1,
			})
			continue
		}

		// --- Weight product (first line: "1   PRODUCT NAME") ---
		// Matches "1   <name>" with no price on this line.
		if strings.HasPrefix(trimmed, "1 ") || strings.HasPrefix(trimmed, "1\t") {
			// Strip the leading "1" and spaces.
			rest := strings.TrimSpace(trimmed[1:])
			// Must not contain a price pattern (two digits comma two digits at end).
			if !regexp.MustCompile(`\d+,\d{2}\s*$`).MatchString(rest) && rest != "" {
				pendingWeightProduct = rest
				continue
			}
		}
	}

	return t, nil
}

// splitLines splits text on newlines, preserving empty lines so the parser can
// detect blank separators.
func splitLines(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}

// parsePrice converts a Spanish-locale price string ("1,99") to float64.
func parsePrice(s string) (float64, error) {
	normalised := strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	return strconv.ParseFloat(normalised, 64)
}
