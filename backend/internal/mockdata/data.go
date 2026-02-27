package mockdata

import (
	"basket-cost/internal/models"
	"strings"
	"time"
)

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// Products contains all mock product data simulating items from digital receipts.
var Products = []models.Product{
	{
		ID:           "1",
		Name:         "LECHE ENTERA HACENDADO 1L",
		Category:     "Lácteos",
		CurrentPrice: 0.89,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 15), Price: 0.79, Store: "Mercadona"},
			{Date: date(2025, 3, 2), Price: 0.82, Store: "Mercadona"},
			{Date: date(2025, 5, 18), Price: 0.85, Store: "Mercadona"},
			{Date: date(2025, 7, 10), Price: 0.85, Store: "Mercadona"},
			{Date: date(2025, 9, 22), Price: 0.89, Store: "Mercadona"},
			{Date: date(2025, 11, 5), Price: 0.89, Store: "Mercadona"},
			{Date: date(2026, 1, 14), Price: 0.89, Store: "Mercadona"},
		},
	},
	{
		ID:           "2",
		Name:         "PAN BIMBO INTEGRAL",
		Category:     "Panadería",
		CurrentPrice: 2.15,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 20), Price: 1.89, Store: "Carrefour"},
			{Date: date(2025, 3, 15), Price: 1.95, Store: "Carrefour"},
			{Date: date(2025, 5, 10), Price: 1.99, Store: "Mercadona"},
			{Date: date(2025, 7, 28), Price: 2.05, Store: "Carrefour"},
			{Date: date(2025, 9, 14), Price: 2.10, Store: "Carrefour"},
			{Date: date(2025, 11, 30), Price: 2.15, Store: "Mercadona"},
			{Date: date(2026, 2, 1), Price: 2.15, Store: "Carrefour"},
		},
	},
	{
		ID:           "3",
		Name:         "ACEITE OLIVA VIRGEN EXTRA CARBONELL 1L",
		Category:     "Aceites",
		CurrentPrice: 8.99,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 5), Price: 7.49, Store: "Mercadona"},
			{Date: date(2025, 2, 18), Price: 7.89, Store: "Carrefour"},
			{Date: date(2025, 4, 10), Price: 8.19, Store: "Mercadona"},
			{Date: date(2025, 6, 22), Price: 8.49, Store: "Lidl"},
			{Date: date(2025, 8, 15), Price: 8.75, Store: "Mercadona"},
			{Date: date(2025, 10, 3), Price: 8.99, Store: "Carrefour"},
			{Date: date(2025, 12, 20), Price: 8.99, Store: "Mercadona"},
			{Date: date(2026, 2, 10), Price: 8.99, Store: "Mercadona"},
		},
	},
	{
		ID:           "4",
		Name:         "HUEVOS CAMPEROS DOCENA",
		Category:     "Huevos",
		CurrentPrice: 2.85,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 12), Price: 2.49, Store: "Mercadona"},
			{Date: date(2025, 3, 25), Price: 2.55, Store: "Lidl"},
			{Date: date(2025, 5, 30), Price: 2.65, Store: "Mercadona"},
			{Date: date(2025, 8, 8), Price: 2.75, Store: "Carrefour"},
			{Date: date(2025, 10, 18), Price: 2.80, Store: "Mercadona"},
			{Date: date(2026, 1, 5), Price: 2.85, Store: "Mercadona"},
		},
	},
	{
		ID:           "5",
		Name:         "ARROZ LARGO SOS 1KG",
		Category:     "Arroces y pastas",
		CurrentPrice: 1.65,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 2, 1), Price: 1.39, Store: "Mercadona"},
			{Date: date(2025, 4, 15), Price: 1.45, Store: "Carrefour"},
			{Date: date(2025, 6, 20), Price: 1.49, Store: "Mercadona"},
			{Date: date(2025, 8, 30), Price: 1.55, Store: "Lidl"},
			{Date: date(2025, 11, 10), Price: 1.59, Store: "Mercadona"},
			{Date: date(2026, 1, 25), Price: 1.65, Store: "Mercadona"},
		},
	},
	{
		ID:           "6",
		Name:         "PASTA ESPAGUETIS GALLO 500G",
		Category:     "Arroces y pastas",
		CurrentPrice: 1.25,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 8), Price: 1.09, Store: "Mercadona"},
			{Date: date(2025, 3, 20), Price: 1.09, Store: "Lidl"},
			{Date: date(2025, 5, 15), Price: 1.15, Store: "Carrefour"},
			{Date: date(2025, 7, 25), Price: 1.19, Store: "Mercadona"},
			{Date: date(2025, 10, 5), Price: 1.25, Store: "Mercadona"},
			{Date: date(2026, 1, 10), Price: 1.25, Store: "Carrefour"},
		},
	},
	{
		ID:           "7",
		Name:         "TOMATE TRITURADO ORLANDO 800G",
		Category:     "Conservas",
		CurrentPrice: 1.49,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 2, 14), Price: 1.19, Store: "Mercadona"},
			{Date: date(2025, 4, 28), Price: 1.25, Store: "Carrefour"},
			{Date: date(2025, 7, 1), Price: 1.29, Store: "Mercadona"},
			{Date: date(2025, 9, 15), Price: 1.39, Store: "Lidl"},
			{Date: date(2025, 11, 28), Price: 1.45, Store: "Mercadona"},
			{Date: date(2026, 2, 5), Price: 1.49, Store: "Mercadona"},
		},
	},
	{
		ID:           "8",
		Name:         "YOGUR NATURAL DANONE PACK 4",
		Category:     "Lácteos",
		CurrentPrice: 1.79,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 22), Price: 1.55, Store: "Carrefour"},
			{Date: date(2025, 3, 10), Price: 1.59, Store: "Mercadona"},
			{Date: date(2025, 5, 25), Price: 1.65, Store: "Mercadona"},
			{Date: date(2025, 8, 3), Price: 1.69, Store: "Lidl"},
			{Date: date(2025, 10, 20), Price: 1.75, Store: "Mercadona"},
			{Date: date(2026, 1, 18), Price: 1.79, Store: "Carrefour"},
		},
	},
	{
		ID:           "9",
		Name:         "POLLO ENTERO FRESCO KG",
		Category:     "Carnes",
		CurrentPrice: 4.25,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 10), Price: 3.69, Store: "Mercadona"},
			{Date: date(2025, 3, 5), Price: 3.79, Store: "Carrefour"},
			{Date: date(2025, 5, 20), Price: 3.89, Store: "Mercadona"},
			{Date: date(2025, 7, 15), Price: 3.99, Store: "Lidl"},
			{Date: date(2025, 9, 28), Price: 4.10, Store: "Mercadona"},
			{Date: date(2025, 12, 10), Price: 4.25, Store: "Mercadona"},
			{Date: date(2026, 2, 15), Price: 4.25, Store: "Carrefour"},
		},
	},
	{
		ID:           "10",
		Name:         "PLATANOS KG",
		Category:     "Frutas",
		CurrentPrice: 1.99,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 18), Price: 1.69, Store: "Mercadona"},
			{Date: date(2025, 3, 30), Price: 1.75, Store: "Lidl"},
			{Date: date(2025, 6, 12), Price: 1.79, Store: "Mercadona"},
			{Date: date(2025, 8, 20), Price: 1.85, Store: "Carrefour"},
			{Date: date(2025, 10, 28), Price: 1.95, Store: "Mercadona"},
			{Date: date(2026, 1, 8), Price: 1.99, Store: "Mercadona"},
		},
	},
	{
		ID:           "11",
		Name:         "MANZANAS GOLDEN KG",
		Category:     "Frutas",
		CurrentPrice: 2.29,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 2, 5), Price: 1.89, Store: "Carrefour"},
			{Date: date(2025, 4, 18), Price: 1.99, Store: "Mercadona"},
			{Date: date(2025, 6, 30), Price: 2.09, Store: "Lidl"},
			{Date: date(2025, 9, 10), Price: 2.15, Store: "Mercadona"},
			{Date: date(2025, 11, 22), Price: 2.25, Store: "Mercadona"},
			{Date: date(2026, 2, 3), Price: 2.29, Store: "Carrefour"},
		},
	},
	{
		ID:           "12",
		Name:         "CERVEZA MAHOU 5 ESTRELLAS PACK 6",
		Category:     "Bebidas",
		CurrentPrice: 4.89,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 25), Price: 4.29, Store: "Mercadona"},
			{Date: date(2025, 4, 5), Price: 4.39, Store: "Carrefour"},
			{Date: date(2025, 6, 18), Price: 4.49, Store: "Mercadona"},
			{Date: date(2025, 8, 25), Price: 4.65, Store: "Lidl"},
			{Date: date(2025, 11, 8), Price: 4.79, Store: "Mercadona"},
			{Date: date(2026, 1, 30), Price: 4.89, Store: "Carrefour"},
		},
	},
	{
		ID:           "13",
		Name:         "CAFE MOLIDO NATURAL MARCILLA 250G",
		Category:     "Desayuno",
		CurrentPrice: 3.49,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 2, 10), Price: 2.89, Store: "Mercadona"},
			{Date: date(2025, 4, 22), Price: 2.99, Store: "Carrefour"},
			{Date: date(2025, 6, 15), Price: 3.15, Store: "Mercadona"},
			{Date: date(2025, 8, 28), Price: 3.25, Store: "Lidl"},
			{Date: date(2025, 11, 15), Price: 3.39, Store: "Mercadona"},
			{Date: date(2026, 2, 8), Price: 3.49, Store: "Mercadona"},
		},
	},
	{
		ID:           "14",
		Name:         "PAPEL HIGIENICO SCOTTEX 12 ROLLOS",
		Category:     "Higiene",
		CurrentPrice: 4.99,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 1, 30), Price: 4.29, Store: "Carrefour"},
			{Date: date(2025, 4, 12), Price: 4.39, Store: "Mercadona"},
			{Date: date(2025, 6, 25), Price: 4.55, Store: "Mercadona"},
			{Date: date(2025, 9, 5), Price: 4.69, Store: "Lidl"},
			{Date: date(2025, 11, 20), Price: 4.85, Store: "Mercadona"},
			{Date: date(2026, 2, 12), Price: 4.99, Store: "Carrefour"},
		},
	},
	{
		ID:           "15",
		Name:         "DETERGENTE SKIP LIQUIDO 30 LAVADOS",
		Category:     "Limpieza",
		CurrentPrice: 7.49,
		PriceHistory: []models.PriceRecord{
			{Date: date(2025, 2, 20), Price: 6.49, Store: "Mercadona"},
			{Date: date(2025, 4, 30), Price: 6.69, Store: "Carrefour"},
			{Date: date(2025, 7, 8), Price: 6.89, Store: "Mercadona"},
			{Date: date(2025, 9, 18), Price: 7.09, Store: "Lidl"},
			{Date: date(2025, 12, 1), Price: 7.29, Store: "Mercadona"},
			{Date: date(2026, 2, 18), Price: 7.49, Store: "Mercadona"},
		},
	},
}

// SearchProducts returns products whose name contains the query string (case-insensitive).
func SearchProducts(query string) []models.SearchResult {
	query = strings.ToLower(query)
	var results []models.SearchResult

	for _, p := range Products {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) {
			minPrice, maxPrice := priceRange(p.PriceHistory)
			results = append(results, models.SearchResult{
				ID:           p.ID,
				Name:         p.Name,
				Category:     p.Category,
				CurrentPrice: p.CurrentPrice,
				MinPrice:     minPrice,
				MaxPrice:     maxPrice,
			})
		}
	}
	return results
}

// GetProductByID returns a product by its ID, or nil if not found.
func GetProductByID(id string) *models.Product {
	for _, p := range Products {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

func priceRange(records []models.PriceRecord) (min, max float64) {
	if len(records) == 0 {
		return 0, 0
	}
	min = records[0].Price
	max = records[0].Price
	for _, r := range records[1:] {
		if r.Price < min {
			min = r.Price
		}
		if r.Price > max {
			max = r.Price
		}
	}
	return min, max
}
