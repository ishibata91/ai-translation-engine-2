package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Parse command line arguments
	dbPath := flag.String("db", "dictionary.db", "Path to the SQLite database file")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: dictionary_builder [options] <xml_file_path>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	xmlFilePath := args[0]


	ctx := context.Background()
	slog.InfoContext(ctx, "Starting Dictionary Builder", "dbPath", *dbPath, "xmlFilePath", xmlFilePath)

	// Open SQLite Database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Use Wire to initialize the Importer
	importer := initializeImporter(db)

	// Initialize the database tables (this is a bit of a manual step here since Importer interface doesn't expose Initialize,
	// ideally Store.Initialize should be called, but since Importer encapsulates Store, we assume the Store handles it or we expose it.
	// For this slice, let's cast back to Store to initialize for the CLI, or we could have a Builder App struct.)
	// To keep it simple, we will execute the schema creation directly here or adjust the Wire setup.
	// We'll just run the initialize logic here manually to ensure tables exist.
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dictionary_entries (
			edid TEXT PRIMARY KEY,
			rec TEXT NOT NULL,
			source TEXT NOT NULL,
			dest TEXT NOT NULL,
			addon TEXT NOT NULL
		);
	`)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize db", "error", err)
		os.Exit(1)
	}

	// Open XML file for reading
	file, err := os.Open(xmlFilePath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to open XML file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// Run the import
	count, err := importer.ImportXML(ctx, file)
	if err != nil {
		slog.ErrorContext(ctx, "Import failed", "error", err)
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Import completed successfully", "imported_count", count)
}
