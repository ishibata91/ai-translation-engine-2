package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/progress"
	dictionary2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/dictionary"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Parse command line arguments
	dbPath := flag.String("db", "dictionary.db", "Path to the SQLite database file")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: dictionary [options] <xml_file_path>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	xmlFilePath := args[0]

	ctx := context.Background()
	slog.InfoContext(ctx, "Starting Dictionary Builder", "db_path", *dbPath, "xml_file_path", xmlFilePath)

	// 1. Initialize DB
	db, dbCleanup, err := datastore.NewSQLiteDB(ctx, *dbPath) // Use *dbPath instead of hardcoded "dictionary.db"
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer dbCleanup()

	// 構築部分
	store, err := dictionary2.NewDictionaryStore(db)
	if err != nil {
		log.Fatalf("failed to initialize dictionary store: %v", err)
	}
	config := dictionary2.DefaultConfig()
	notifier := progress.NewNoopNotifier()
	logger := slog.Default() // Define logger as it's used in NewImporter
	importer := dictionary2.NewImporter(config, store, notifier, logger)

	// Open XML file for reading
	file, err := os.Open(xmlFilePath) // Use xmlFilePath
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Run the import
	count, err := importer.ImportXML(ctx, 1, xmlFilePath, file) // Use xmlFilePath
	if err != nil {
		slog.ErrorContext(ctx, "Import failed", "error", err)
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Import completed successfully", "imported_count", count)
}
