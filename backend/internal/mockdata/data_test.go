package mockdata_test

import (
	"basket-cost/internal/mockdata"
	"testing"
)

func TestSearchProducts_EmptyQuery_ReturnsAll(t *testing.T) {
	results := mockdata.SearchProducts("")
	if len(results) != len(mockdata.Products) {
		t.Errorf("expected %d results, got %d", len(mockdata.Products), len(results))
	}
}

func TestSearchProducts_MatchingQuery(t *testing.T) {
	results := mockdata.SearchProducts("leche")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'leche', got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Errorf("expected ID '1', got '%s'", results[0].ID)
	}
}

func TestSearchProducts_CaseInsensitive(t *testing.T) {
	lower := mockdata.SearchProducts("leche")
	upper := mockdata.SearchProducts("LECHE")
	mixed := mockdata.SearchProducts("Leche")
	if len(lower) != len(upper) || len(lower) != len(mixed) {
		t.Error("search should be case-insensitive")
	}
}

func TestSearchProducts_NoMatch_ReturnsEmpty(t *testing.T) {
	results := mockdata.SearchProducts("xyznonexistentproduct")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchProducts_MinMaxPriceCorrect(t *testing.T) {
	results := mockdata.SearchProducts("leche entera hacendado")
	if len(results) == 0 {
		t.Fatal("milk product not found")
	}
	r := results[0]
	if r.MinPrice != 0.79 {
		t.Errorf("expected minPrice 0.79, got %f", r.MinPrice)
	}
	if r.MaxPrice != 0.89 {
		t.Errorf("expected maxPrice 0.89, got %f", r.MaxPrice)
	}
}

func TestSearchProducts_ResultHasRequiredFields(t *testing.T) {
	results := mockdata.SearchProducts("pan")
	if len(results) == 0 {
		t.Fatal("no results found for 'pan'")
	}
	r := results[0]
	if r.ID == "" {
		t.Error("ID should not be empty")
	}
	if r.Name == "" {
		t.Error("Name should not be empty")
	}
	if r.CurrentPrice <= 0 {
		t.Error("CurrentPrice should be greater than 0")
	}
}

func TestGetProductByID_Exists(t *testing.T) {
	p := mockdata.GetProductByID("1")
	if p == nil {
		t.Fatal("expected product with ID '1', got nil")
	}
	if p.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", p.ID)
	}
	if p.Name == "" {
		t.Error("Name should not be empty")
	}
	if len(p.PriceHistory) == 0 {
		t.Error("PriceHistory should not be empty")
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	p := mockdata.GetProductByID("9999")
	if p != nil {
		t.Errorf("expected nil for nonexistent ID, got %+v", p)
	}
}

func TestGetProductByID_EmptyID(t *testing.T) {
	p := mockdata.GetProductByID("")
	if p != nil {
		t.Errorf("expected nil for empty ID, got %+v", p)
	}
}

func TestGetProductByID_AllProductsRetrievable(t *testing.T) {
	for _, product := range mockdata.Products {
		p := mockdata.GetProductByID(product.ID)
		if p == nil {
			t.Errorf("product with ID '%s' not found", product.ID)
		}
	}
}
