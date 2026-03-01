package main

import (
	"context"
	"embed"
	"log"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/task"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. Initialize DB (config.db)
	db, dbCleanup, err := datastore.NewSQLiteDB("config.db")
	if err != nil {
		log.Fatalf("failed to initialize config database: %v", err)
	}
	defer dbCleanup()

	// 2. Run Migrations
	if err := config.Migrate(context.Background(), db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	// 3. Setup Task Manager
	logger := telemetry.ProvideLogger()
	taskStore := task.NewStore(db)
	taskManager := task.NewManager(context.TODO(), logger, taskStore) // Context will be set in startup

	// 4. Setup Bridge
	taskBridge := task.NewBridge(taskManager)

	// 5. Setup Dictionary System (dictionary.db)
	dictDB, dictDBCleanup, err := datastore.NewSQLiteDB("dictionary.db")
	if err != nil {
		log.Fatalf("failed to initialize dictionary database: %v", err)
	}
	defer dictDBCleanup()

	dictStore, err := dictionary.NewDictionaryStore(dictDB)
	if err != nil {
		log.Fatalf("failed to initialize dictionary store: %v", err)
	}
	wailsNotifier := progress.NewWailsNotifier(logger)
	wailsNotifier.SetEventName("dictionary:import_progress") // 必要に応じて変更可能

	// Wails AST needs `New<StructName>` pattern or `&StructName{}` pattern to discover bindings properly
	dictConfig := dictionary.DefaultConfig()
	dictImporter := dictionary.NewImporter(dictConfig, dictStore, wailsNotifier, logger)
	dictService := dictionary.NewDictionaryService(dictStore, dictImporter, logger)

	// Create an instance of the app structure
	app := NewApp()
	app.SetDictService(dictService)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "AI Translation Engine",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			// Pass Wails context to TaskManager and Initialize it
			taskManager.SetContext(ctx)
			if err := taskManager.Initialize(ctx); err != nil {
				log.Printf("failed to initialize task manager: %v", err)
			}
			// Dictionary service progress notifier
			wailsNotifier.SetContext(ctx)
		},
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
			taskBridge,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
