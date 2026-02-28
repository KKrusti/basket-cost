package enricher

import (
	"context"
	"fmt"
	"log"

	"basket-cost/internal/store"
)

// minMatchScore is the minimum fraction of local keywords that must match a
// Mercadona product's keyword set for the match to be accepted.
// A value of 1.0 means all local keywords must appear in the API entry.
const minMatchScore = 1.0

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
// using keyword scoring, and updates image_url for every match.
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
		localKW := keywords(normalise(p.Name))
		if len(localKW) == 0 {
			res.Skipped++
			continue
		}

		url, ok := bestMatch(localKW, index)
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

// bestMatch finds the ProductEntry whose keyword set covers the most local
// keywords. It returns the thumbnail URL and true if the match score meets
// minMatchScore, otherwise returns ("", false).
func bestMatch(localKW []string, index ProductIndex) (string, bool) {
	bestScore := 0.0
	bestURL := ""

	for _, entry := range index {
		// Build a set from the entry's keywords for O(1) lookup.
		entrySet := make(map[string]bool, len(entry.Keywords))
		for _, k := range entry.Keywords {
			entrySet[k] = true
		}

		matched := 0
		for _, k := range localKW {
			if entrySet[k] {
				matched++
			}
		}

		score := float64(matched) / float64(len(localKW))
		if score > bestScore {
			bestScore = score
			bestURL = entry.Thumbnail
		}
	}

	if bestScore >= minMatchScore {
		return bestURL, true
	}
	return "", false
}
