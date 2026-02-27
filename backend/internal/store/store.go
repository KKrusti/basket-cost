// Package store provides the data access layer backed by SQLite.
package store

import (
	"basket-cost/internal/models"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Store is the interface the HTTP handlers depend on.
// Both the real SQLite implementation and test fakes satisfy it.
type Store interface {
	SearchProducts(query string) ([]models.SearchResult, error)
	GetProductByID(id string) (*models.Product, error)
	InsertProduct(p models.Product) error
	// UpsertPriceRecord ensures the named product exists (creating it if needed)
	// and appends a new price record for the given observation.
	UpsertPriceRecord(name string, record models.PriceRecord) error
	// UpdateProductImageURL sets the image URL for the product with the given ID.
	UpdateProductImageURL(id, imageURL string) error
}

// SQLiteStore is the production Store backed by a *sql.DB.
type SQLiteStore struct {
	db *sql.DB
}

// New returns a SQLiteStore wrapping the given database connection.
func New(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// InsertProduct inserts a product and all its price records inside a single
// transaction. If the product already exists it is skipped (idempotent seed).
func (s *SQLiteStore) InsertProduct(p models.Product) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(
		`INSERT OR IGNORE INTO products (id, name, category) VALUES (?, ?, ?)`,
		p.ID, p.Name, p.Category,
	)
	if err != nil {
		return fmt.Errorf("insert product %s: %w", p.ID, err)
	}

	for _, r := range p.PriceHistory {
		_, err = tx.Exec(
			`INSERT INTO price_records (product_id, date, price, store) VALUES (?, ?, ?, ?)`,
			p.ID, r.Date.Format(time.DateOnly), r.Price, r.Store,
		)
		if err != nil {
			return fmt.Errorf("insert price record for product %s: %w", p.ID, err)
		}
	}

	return tx.Commit()
}

// SearchProducts returns products whose name contains query (case-insensitive).
// An empty query returns all products.
func (s *SQLiteStore) SearchProducts(query string) ([]models.SearchResult, error) {
	const baseSQL = `
		SELECT
			p.id,
			p.name,
			p.category,
			p.image_url,
			(SELECT price FROM price_records WHERE product_id = p.id ORDER BY date DESC LIMIT 1) AS current_price,
			(SELECT MIN(price) FROM price_records WHERE product_id = p.id)                        AS min_price,
			(SELECT MAX(price) FROM price_records WHERE product_id = p.id)                        AS max_price
		FROM products p
	`

	var (
		rows *sql.Rows
		err  error
	)

	if strings.TrimSpace(query) == "" {
		rows, err = s.db.Query(baseSQL + " ORDER BY p.name")
	} else {
		rows, err = s.db.Query(
			baseSQL+` WHERE p.name LIKE ? ORDER BY p.name`,
			"%"+query+"%",
		)
	}
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var r models.SearchResult
		var category, imageURL sql.NullString
		var currentPrice, minPrice, maxPrice sql.NullFloat64
		if err := rows.Scan(&r.ID, &r.Name, &category, &imageURL, &currentPrice, &minPrice, &maxPrice); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		r.Category = category.String
		r.ImageURL = imageURL.String
		r.CurrentPrice = currentPrice.Float64
		r.MinPrice = minPrice.Float64
		r.MaxPrice = maxPrice.Float64
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	if results == nil {
		results = []models.SearchResult{}
	}
	return results, nil
}

// reNonAlphanumeric matches any character that is not a lowercase letter or digit.
var reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a product name to a stable, URL-safe ID.
// Example: "LECHE ENTERA HACENDADO 1L" → "leche-entera-hacendado-1l"
func slugify(name string) string {
	lower := strings.ToLower(name)
	slug := reNonAlphanumeric.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// UpsertPriceRecord ensures a product with the given name exists in the
// database (creating it with a generated ID if necessary) and then inserts a
// new price record for it.
func (s *SQLiteStore) UpsertPriceRecord(name string, record models.PriceRecord) error {
	id := slugify(name)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Insert product if it does not exist yet.
	_, err = tx.Exec(
		`INSERT OR IGNORE INTO products (id, name, category) VALUES (?, ?, ?)`,
		id, name, "",
	)
	if err != nil {
		return fmt.Errorf("upsert product %q: %w", name, err)
	}

	// Insert the price record.
	_, err = tx.Exec(
		`INSERT INTO price_records (product_id, date, price, store) VALUES (?, ?, ?, ?)`,
		id, record.Date.Format(time.DateOnly), record.Price, record.Store,
	)
	if err != nil {
		return fmt.Errorf("insert price record for product %q: %w", name, err)
	}

	return tx.Commit()
}

// GetProductByID returns the full product with its price history, or nil if not found.
func (s *SQLiteStore) GetProductByID(id string) (*models.Product, error) {
	row := s.db.QueryRow(
		`SELECT id, name, category, image_url FROM products WHERE id = ?`, id,
	)

	var p models.Product
	var category, imageURL sql.NullString
	if err := row.Scan(&p.ID, &p.Name, &category, &imageURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get product %s: %w", id, err)
	}
	p.Category = category.String
	p.ImageURL = imageURL.String

	rows, err := s.db.Query(
		`SELECT date, price, store FROM price_records WHERE product_id = ? ORDER BY date ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get price records for product %s: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var rec models.PriceRecord
		var dateStr string
		if err := rows.Scan(&dateStr, &rec.Price, &rec.Store); err != nil {
			return nil, fmt.Errorf("scan price record: %w", err)
		}
		rec.Date, err = time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateStr, err)
		}
		p.PriceHistory = append(p.PriceHistory, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price records: %w", err)
	}

	// Derive CurrentPrice from the most recent record.
	if len(p.PriceHistory) > 0 {
		p.CurrentPrice = p.PriceHistory[len(p.PriceHistory)-1].Price
	}

	return &p, nil
}

// UpdateProductImageURL sets the image_url for the product with the given ID.
// It is a no-op if no product with that ID exists.
func (s *SQLiteStore) UpdateProductImageURL(id, imageURL string) error {
	_, err := s.db.Exec(
		`UPDATE products SET image_url = ? WHERE id = ?`, imageURL, id,
	)
	if err != nil {
		return fmt.Errorf("update image_url for product %s: %w", id, err)
	}
	return nil
}
