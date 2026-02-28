package store_test

import (
	"testing"
	"time"

	"basket-cost/internal/database"
	"basket-cost/internal/models"
	"basket-cost/internal/store"
)

// newTestStore opens an in-memory SQLite DB, applies migrations, and returns a
// ready-to-use SQLiteStore. The caller does not need to close it (test cleanup
// handles the underlying *sql.DB via t.Cleanup).
func newTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.New(db)
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// sampleProduct returns a deterministic product for use across tests.
func sampleProduct(id string) models.Product {
	return models.Product{
		ID:       id,
		Name:     "LECHE ENTERA HACENDADO 1L",
		Category: "Lácteos",
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 10), Price: 0.79, Store: "Mercadona"},
			{Date: date(2025, 6, 15), Price: 0.85, Store: "Mercadona"},
			{Date: date(2026, 1, 20), Price: 0.89, Store: "Mercadona"},
		},
	}
}

// ---------- InsertProduct ----------

func TestInsertProduct_Success(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("1")
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("InsertProduct returned unexpected error: %v", err)
	}
}

func TestInsertProduct_Idempotent(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("1")

	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	// Second insert of same product must not error (INSERT OR IGNORE).
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("second insert (idempotent): %v", err)
	}
}

func TestInsertProduct_PriceHistoryPersisted(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("42")

	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("InsertProduct: %v", err)
	}

	got, err := s.GetProductByID("42")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected product, got nil")
	}
	if len(got.PriceHistory) != len(p.PriceHistory) {
		t.Errorf("price history length: want %d, got %d", len(p.PriceHistory), len(got.PriceHistory))
	}
}

// ---------- GetProductByID ----------

func TestGetProductByID_NotFound(t *testing.T) {
	s := newTestStore(t)
	got, err := s.GetProductByID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing product, got %+v", got)
	}
}

func TestGetProductByID_Fields(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("7")
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := s.GetProductByID("7")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected product, got nil")
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ID", got.ID, p.ID},
		{"Name", got.Name, p.Name},
		{"Category", got.Category, p.Category},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("want %q, got %q", tt.want, tt.got)
			}
		})
	}
}

func TestGetProductByID_CurrentPriceDerivedFromLatestRecord(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("5")
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := s.GetProductByID("5")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	// Latest record (2026-01-20) has Price 0.89.
	if got.CurrentPrice != 0.89 {
		t.Errorf("CurrentPrice: want 0.89, got %f", got.CurrentPrice)
	}
}

func TestGetProductByID_PriceHistoryOrderedByDate(t *testing.T) {
	s := newTestStore(t)
	// Insert records deliberately out of order.
	p := models.Product{
		ID:       "99",
		Name:     "TEST PRODUCT",
		Category: "Test",
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 6, 1), Price: 2.00, Store: "A"},
			{Date: date(2025, 1, 1), Price: 1.00, Store: "A"},
			{Date: date(2025, 12, 1), Price: 3.00, Store: "A"},
		},
	}
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, err := s.GetProductByID("99")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}

	for i := 1; i < len(got.PriceHistory); i++ {
		if got.PriceHistory[i].Date.Before(got.PriceHistory[i-1].Date) {
			t.Errorf("price history not ordered by date at index %d", i)
		}
	}
}

// ---------- SearchProducts ----------

func TestSearchProducts_EmptyQuery_ReturnsAll(t *testing.T) {
	s := newTestStore(t)

	products := []models.Product{
		{ID: "a", Name: "LECHE ENTERA", Category: "Lácteos", PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 0.89, Store: "Mercadona"},
		}},
		{ID: "b", Name: "PAN INTEGRAL", Category: "Panadería", PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 1.25, Store: "Mercadona"},
		}},
	}
	for _, p := range products {
		if err := s.InsertProduct(p); err != nil {
			t.Fatalf("insert %s: %v", p.ID, err)
		}
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("want 2 results, got %d", len(results))
	}
}

func TestSearchProducts_WithQuery_ReturnsMatches(t *testing.T) {
	s := newTestStore(t)

	products := []models.Product{
		{ID: "a", Name: "LECHE ENTERA HACENDADO", Category: "Lácteos", PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 0.89, Store: "Mercadona"},
		}},
		{ID: "b", Name: "PAN INTEGRAL BIMBO", Category: "Panadería", PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 1.25, Store: "Mercadona"},
		}},
	}
	for _, p := range products {
		if err := s.InsertProduct(p); err != nil {
			t.Fatalf("insert %s: %v", p.ID, err)
		}
	}

	results, err := s.SearchProducts("leche")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("want 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "a" {
		t.Errorf("want product 'a', got %q", results[0].ID)
	}
}

func TestSearchProducts_NoMatch_ReturnsEmptySlice(t *testing.T) {
	s := newTestStore(t)
	if err := s.InsertProduct(sampleProduct("1")); err != nil {
		t.Fatalf("insert: %v", err)
	}

	results, err := s.SearchProducts("xyznonexistent")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if results == nil {
		t.Error("want empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

func TestSearchProducts_CaseInsensitive(t *testing.T) {
	s := newTestStore(t)
	p := models.Product{
		ID:       "ci",
		Name:     "LECHE ENTERA",
		Category: "Lácteos",
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 0.89, Store: "Mercadona"},
		},
	}
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	for _, q := range []string{"leche", "LECHE", "Leche", "lEcHe"} {
		t.Run(q, func(t *testing.T) {
			results, err := s.SearchProducts(q)
			if err != nil {
				t.Fatalf("SearchProducts(%q): %v", q, err)
			}
			if len(results) != 1 {
				t.Errorf("query %q: want 1 result, got %d", q, len(results))
			}
		})
	}
}

func TestSearchProducts_MinMaxPrice(t *testing.T) {
	s := newTestStore(t)
	p := models.Product{
		ID:       "mm",
		Name:     "ACEITE OLIVA",
		Category: "Aceites",
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 5.00, Store: "A"},
			{Date: date(2025, 6, 1), Price: 8.00, Store: "A"},
			{Date: date(2025, 12, 1), Price: 6.50, Store: "A"},
		},
	}
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	r := results[0]
	if r.MinPrice != 5.00 {
		t.Errorf("MinPrice: want 5.00, got %f", r.MinPrice)
	}
	if r.MaxPrice != 8.00 {
		t.Errorf("MaxPrice: want 8.00, got %f", r.MaxPrice)
	}
}

func TestSearchProducts_LastPurchaseDatePopulated(t *testing.T) {
	s := newTestStore(t)
	p := models.Product{
		ID:       "lpd",
		Name:     "YOGUR NATURAL",
		Category: "Lácteos",
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 1), Price: 0.30, Store: "Mercadona"},
			{Date: date(2025, 9, 15), Price: 0.35, Store: "Mercadona"},
		},
	}
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].LastPurchaseDate != "2025-09-15" {
		t.Errorf("LastPurchaseDate: want %q, got %q", "2025-09-15", results[0].LastPurchaseDate)
	}
}

func TestSearchProducts_OrderedByLastPurchaseDateDesc(t *testing.T) {
	s := newTestStore(t)

	// Insert two products with different last purchase dates.
	older := models.Product{
		ID:   "old",
		Name: "PAN INTEGRAL",
		PriceHistory: []models.PriceRecord{
			{Date: date(2024, 1, 1), Price: 1.00, Store: "A"},
		},
	}
	newer := models.Product{
		ID:   "new",
		Name: "LECHE ENTERA",
		PriceHistory: []models.PriceRecord{
			{Date: date(2026, 2, 1), Price: 0.89, Store: "A"},
		},
	}
	if err := s.InsertProduct(older); err != nil {
		t.Fatalf("insert older: %v", err)
	}
	if err := s.InsertProduct(newer); err != nil {
		t.Fatalf("insert newer: %v", err)
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	// Most recently purchased product should be first.
	if results[0].ID != "new" {
		t.Errorf("first result: want %q (newest), got %q", "new", results[0].ID)
	}
	if results[1].ID != "old" {
		t.Errorf("second result: want %q (oldest), got %q", "old", results[1].ID)
	}
}

func TestSearchProducts_EmptyDB_ReturnsEmptySlice(t *testing.T) {
	s := newTestStore(t)
	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts on empty DB: %v", err)
	}
	if results == nil {
		t.Error("want empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

// ---------- UpdateProductImageURL ----------

func TestUpdateProductImageURL_SetsURL(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("img-test")
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	const url = "https://prod-mercadona.imgix.net/images/abc123.jpg?fit=crop&h=300&w=300"
	if err := s.UpdateProductImageURL("img-test", url); err != nil {
		t.Fatalf("UpdateProductImageURL: %v", err)
	}

	got, err := s.GetProductByID("img-test")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	if got.ImageURL != url {
		t.Errorf("ImageURL: want %q, got %q", url, got.ImageURL)
	}
}

func TestUpdateProductImageURL_SearchResultIncludesURL(t *testing.T) {
	s := newTestStore(t)
	p := sampleProduct("img-search")
	if err := s.InsertProduct(p); err != nil {
		t.Fatalf("insert: %v", err)
	}

	const url = "https://prod-mercadona.imgix.net/images/xyz.jpg?fit=crop&h=300&w=300"
	if err := s.UpdateProductImageURL("img-search", url); err != nil {
		t.Fatalf("UpdateProductImageURL: %v", err)
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].ImageURL != url {
		t.Errorf("SearchResult.ImageURL: want %q, got %q", url, results[0].ImageURL)
	}
}

func TestUpdateProductImageURL_NoOpOnMissingProduct(t *testing.T) {
	s := newTestStore(t)
	// Must not return an error even if the product does not exist.
	if err := s.UpdateProductImageURL("nonexistent", "http://example.com/img.jpg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- IsFileProcessed / MarkFileProcessed ----------

func TestIsFileProcessed_UnknownFile_ReturnsFalse(t *testing.T) {
	s := newTestStore(t)
	got, err := s.IsFileProcessed("ticket-2026-01.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false for unknown file, got true")
	}
}

func TestMarkFileProcessed_ThenIsFileProcessed_ReturnsTrue(t *testing.T) {
	s := newTestStore(t)
	filename := "ticket-2026-02.pdf"

	if err := s.MarkFileProcessed(filename, date(2026, 2, 1)); err != nil {
		t.Fatalf("MarkFileProcessed: %v", err)
	}

	got, err := s.IsFileProcessed(filename)
	if err != nil {
		t.Fatalf("IsFileProcessed: %v", err)
	}
	if !got {
		t.Error("expected true after marking as processed, got false")
	}
}

func TestMarkFileProcessed_Idempotent(t *testing.T) {
	s := newTestStore(t)
	filename := "ticket-2026-03.pdf"

	if err := s.MarkFileProcessed(filename, date(2026, 3, 1)); err != nil {
		t.Fatalf("first MarkFileProcessed: %v", err)
	}
	// Second call with same filename must not return an error (INSERT OR IGNORE).
	if err := s.MarkFileProcessed(filename, date(2026, 3, 2)); err != nil {
		t.Fatalf("second MarkFileProcessed (idempotent): %v", err)
	}
}

func TestIsFileProcessed_DifferentFilenames_IndependentTracking(t *testing.T) {
	s := newTestStore(t)

	if err := s.MarkFileProcessed("a.pdf", date(2026, 1, 1)); err != nil {
		t.Fatalf("MarkFileProcessed a.pdf: %v", err)
	}

	gotA, err := s.IsFileProcessed("a.pdf")
	if err != nil {
		t.Fatalf("IsFileProcessed a.pdf: %v", err)
	}
	gotB, err := s.IsFileProcessed("b.pdf")
	if err != nil {
		t.Fatalf("IsFileProcessed b.pdf: %v", err)
	}

	if !gotA {
		t.Error("a.pdf: expected true, got false")
	}
	if gotB {
		t.Error("b.pdf: expected false, got true")
	}
}

// ---------- UpsertPriceRecordBatch ----------

func TestUpsertPriceRecordBatch_AllEntriesCommitted(t *testing.T) {
	s := newTestStore(t)

	entries := []models.PriceRecordEntry{
		{Name: "LECHE ENTERA", Record: models.PriceRecord{Date: date(2026, 1, 1), Price: 0.89, Store: "Mercadona"}},
		{Name: "PAN INTEGRAL", Record: models.PriceRecord{Date: date(2026, 1, 1), Price: 1.25, Store: "Mercadona"}},
		{Name: "YOGUR NATURAL", Record: models.PriceRecord{Date: date(2026, 1, 1), Price: 0.35, Store: "Mercadona"}},
	}
	if err := s.UpsertPriceRecordBatch(entries); err != nil {
		t.Fatalf("UpsertPriceRecordBatch: %v", err)
	}

	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("want 3 products, got %d", len(results))
	}
}

func TestUpsertPriceRecordBatch_EmptySlice_NoOp(t *testing.T) {
	s := newTestStore(t)
	if err := s.UpsertPriceRecordBatch([]models.PriceRecordEntry{}); err != nil {
		t.Fatalf("UpsertPriceRecordBatch with empty slice: %v", err)
	}
	results, err := s.SearchProducts("")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 products after no-op batch, got %d", len(results))
	}
}

func TestUpsertPriceRecordBatch_IdempotentProduct(t *testing.T) {
	s := newTestStore(t)

	entry := models.PriceRecordEntry{
		Name:   "LECHE ENTERA",
		Record: models.PriceRecord{Date: date(2026, 1, 1), Price: 0.89, Store: "Mercadona"},
	}
	// Insert twice: the product row should appear only once, but two price records.
	if err := s.UpsertPriceRecordBatch([]models.PriceRecordEntry{entry}); err != nil {
		t.Fatalf("first batch: %v", err)
	}
	entry2 := models.PriceRecordEntry{
		Name:   "LECHE ENTERA",
		Record: models.PriceRecord{Date: date(2026, 2, 1), Price: 0.92, Store: "Mercadona"},
	}
	if err := s.UpsertPriceRecordBatch([]models.PriceRecordEntry{entry2}); err != nil {
		t.Fatalf("second batch: %v", err)
	}

	p, err := s.GetProductByID("leche-entera")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	if p == nil {
		t.Fatal("expected product, got nil")
	}
	if len(p.PriceHistory) != 2 {
		t.Errorf("want 2 price records, got %d", len(p.PriceHistory))
	}
}

func TestUpsertPriceRecordBatch_PriceAndDatePreserved(t *testing.T) {
	s := newTestStore(t)

	want := models.PriceRecord{Date: date(2026, 3, 15), Price: 2.49, Store: "Mercadona"}
	entries := []models.PriceRecordEntry{
		{Name: "ACEITE GIRASOL", Record: want},
	}
	if err := s.UpsertPriceRecordBatch(entries); err != nil {
		t.Fatalf("UpsertPriceRecordBatch: %v", err)
	}

	p, err := s.GetProductByID("aceite-girasol")
	if err != nil {
		t.Fatalf("GetProductByID: %v", err)
	}
	if p == nil || len(p.PriceHistory) == 0 {
		t.Fatal("expected product with price history")
	}
	got := p.PriceHistory[0]
	if got.Price != want.Price {
		t.Errorf("Price: want %.2f, got %.2f", want.Price, got.Price)
	}
	if !got.Date.Equal(want.Date) {
		t.Errorf("Date: want %s, got %s", want.Date, got.Date)
	}
	if got.Store != want.Store {
		t.Errorf("Store: want %q, got %q", want.Store, got.Store)
	}
}

// ---------- GetProductsWithoutImage ----------

func TestGetProductsWithoutImage_ReturnsOnlyUnimaged(t *testing.T) {
	s := newTestStore(t)

	// Insert two products via UpsertPriceRecord; neither has an image yet.
	rec := models.PriceRecord{Date: date(2025, 1, 1), Price: 1.00, Store: "Mercadona"}
	if err := s.UpsertPriceRecord("LECHE ENTERA", rec); err != nil {
		t.Fatalf("upsert leche: %v", err)
	}
	if err := s.UpsertPriceRecord("PAN MOLDE", rec); err != nil {
		t.Fatalf("upsert pan: %v", err)
	}

	// Give one of them an image.
	if err := s.UpdateProductImageURL("leche-entera", "https://img/leche.jpg"); err != nil {
		t.Fatalf("update image: %v", err)
	}

	got, err := s.GetProductsWithoutImage()
	if err != nil {
		t.Fatalf("GetProductsWithoutImage: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 product without image, got %d", len(got))
	}
	if got[0].ID != "pan-molde" {
		t.Errorf("expected pan-molde, got %q", got[0].ID)
	}
	if got[0].Name != "PAN MOLDE" {
		t.Errorf("expected name PAN MOLDE, got %q", got[0].Name)
	}
}

func TestGetProductsWithoutImage_EmptyWhenAllHaveImage(t *testing.T) {
	s := newTestStore(t)

	rec := models.PriceRecord{Date: date(2025, 1, 1), Price: 1.00, Store: "Mercadona"}
	if err := s.UpsertPriceRecord("LECHE ENTERA", rec); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := s.UpdateProductImageURL("leche-entera", "https://img/leche.jpg"); err != nil {
		t.Fatalf("update image: %v", err)
	}

	got, err := s.GetProductsWithoutImage()
	if err != nil {
		t.Fatalf("GetProductsWithoutImage: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d products", len(got))
	}
}

func TestGetProductsWithoutImage_EmptyStore(t *testing.T) {
	s := newTestStore(t)
	got, err := s.GetProductsWithoutImage()
	if err != nil {
		t.Fatalf("GetProductsWithoutImage: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice on empty store, got %d", len(got))
	}
}
