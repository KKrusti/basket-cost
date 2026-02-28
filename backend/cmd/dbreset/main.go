// cmd/dbreset borra todos los datos de la base de datos SQLite (productos, precios e historial de tickets).
package main

import (
	"database/sql"
	"flag"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := flag.String("db", "basket-cost.db", "ruta al fichero SQLite")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	tables := []string{"price_records", "processed_files", "products"}
	for _, t := range tables {
		if _, err := db.Exec("DELETE FROM " + t); err != nil {
			log.Fatalf("delete from %s: %v", t, err)
		}
		log.Printf("tabla %s vaciada", t)
	}

	if _, err := db.Exec("VACUUM"); err != nil {
		log.Fatalf("vacuum: %v", err)
	}
	log.Println("Base de datos vaciada y compactada.")
}
