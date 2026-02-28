package models

import "time"

// PriceRecord represents a single price observation for a product,
// typically extracted from a digital receipt/ticket.
type PriceRecord struct {
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
	Store string    `json:"store,omitempty"`
}

// Product represents a grocery item with its price history.
type Product struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Category     string        `json:"category,omitempty"`
	ImageURL     string        `json:"imageUrl,omitempty"`
	CurrentPrice float64       `json:"currentPrice"`
	PriceHistory []PriceRecord `json:"priceHistory"`
}

// SearchResult is a lightweight version of Product returned in search listings.
type SearchResult struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category,omitempty"`
	ImageURL     string  `json:"imageUrl,omitempty"`
	CurrentPrice float64 `json:"currentPrice"`
	MinPrice     float64 `json:"minPrice"`
	MaxPrice     float64 `json:"maxPrice"`
}

// PriceRecordEntry is the unit of work for batch price-record persistence.
// It pairs a product name with the price observation to record.
type PriceRecordEntry struct {
	Name   string
	Record PriceRecord
}
