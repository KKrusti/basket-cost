// Command seed populates the database by importing Mercadona PDF receipts.
//
// Usage:
//
//	go run ./cmd/seed [flags] <pdf-file> [<pdf-file> ...]
//
// Flags:
//
//	-db string   path to the SQLite database file (default "basket-cost.db")
//	-dir string  directory containing PDF files to import (processed before positional args)
package main

import (
	"basket-cost/internal/database"
	"basket-cost/internal/store"
	"basket-cost/internal/ticket"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dbPath := flag.String("db", "basket-cost.db", "path to the SQLite database file")
	dirPath := flag.String("dir", "", "directory of PDF files to import")
	flag.Parse()

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	s := store.New(db)
	imp := ticket.NewImporter(ticket.NewExtractor(), ticket.NewMercadonaParser(), s)

	// Collect PDF paths: -dir first, then positional arguments.
	var paths []string
	if *dirPath != "" {
		entries, err := os.ReadDir(*dirPath)
		if err != nil {
			log.Fatalf("read dir %q: %v", *dirPath, err)
		}
		for _, e := range entries {
			if e.Type()&fs.ModeSymlink != 0 {
				continue
			}
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".pdf") {
				paths = append(paths, filepath.Join(*dirPath, e.Name()))
			}
		}
	}
	paths = append(paths, flag.Args()...)

	if len(paths) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var totalImported int
	var totalErrors int

	for _, p := range paths {
		result, importErr := importFile(imp, p)
		if result != nil {
			totalImported += result.LinesImported
			fmt.Printf("OK  %-50s  invoice=%s  lines=%d\n",
				filepath.Base(p), result.InvoiceNumber, result.LinesImported)
		}
		if importErr != nil {
			totalErrors++
			fmt.Fprintf(os.Stderr, "ERR %-50s  %v\n", filepath.Base(p), importErr)
		}
	}

	fmt.Printf("\n--- Resultado ---\n")
	fmt.Printf("Archivos procesados : %d\n", len(paths))
	fmt.Printf("Líneas importadas   : %d\n", totalImported)
	fmt.Printf("Errores             : %d\n", totalErrors)

	if totalErrors > 0 {
		os.Exit(1)
	}
}

// importFile opens a PDF from disk and delegates to the Importer.
func importFile(imp *ticket.Importer, path string) (*ticket.ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return imp.Import(f, info.Size())
}
