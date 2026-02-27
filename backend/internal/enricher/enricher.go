package enricher

import (
	"context"
	"fmt"
	"log"

	"basket-cost/internal/store"
)

// Enricher downloads the Mercadona product catalogue and updates image URLs
// for matching products in the local store.
type Enricher struct {
	client *MercadonaClient
	store  store.Store
}

// New returns an Enricher backed by the given store.
func New(s store.Store) *Enricher {
	return &Enricher{
		client: NewMercadonaClient(),
		store:  s,
	}
}

// EnrichResult summarises the outcome of a single enrichment run.
type EnrichResult struct {
	Total   int // products inspected
	Updated int // products whose image URL was set
	Skipped int // products with no match in the Mercadona index
}

// Run fetches the Mercadona catalogue, matches it against the local products
// by normalised name, and updates image_url for every match.
func (e *Enricher) Run(ctx context.Context) (EnrichResult, error) {
	log.Println("enricher: building Mercadona product index…")
	index, err := e.client.BuildProductIndex(ctx)
	if err != nil {
		return EnrichResult{}, fmt.Errorf("build product index: %w", err)
	}
	log.Printf("enricher: index contains %d entries", len(index))

	results, err := e.store.SearchProducts("")
	if err != nil {
		return EnrichResult{}, fmt.Errorf("list products: %w", err)
	}

	var res EnrichResult
	res.Total = len(results)

	for _, p := range results {
		key := normalise(p.Name)
		url, ok := index[key]
		if !ok {
			res.Skipped++
			continue
		}
		if err := e.store.UpdateProductImageURL(p.ID, url); err != nil {
			return res, fmt.Errorf("update image for %s: %w", p.ID, err)
		}
		res.Updated++
	}

	return res, nil
}
