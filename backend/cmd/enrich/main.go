// Command enrich fetches product images from the Mercadona public API and
// persists them in the local SQLite database.
//
// Usage:
//
//	go run ./cmd/enrich/main.go -db <path-to-db>
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"basket-cost/internal/database"
	"basket-cost/internal/enricher"
	"basket-cost/internal/store"
)

func main() {
	dbPath := flag.String("db", "basket-cost.db", "path to the SQLite database file")
	flag.Parse()

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database %q: %v", *dbPath, err)
	}
	defer db.Close()

	s := store.New(db)
	e := enricher.New(s)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	result, err := e.Run(ctx)
	if err != nil {
		log.Fatalf("enricher run: %v", err)
	}

	log.Printf("enricher done — total: %d, updated: %d, skipped: %d",
		result.Total, result.Updated, result.Skipped)
}
