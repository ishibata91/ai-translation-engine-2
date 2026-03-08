package main

import (
	"context"
	"embed"
	"log"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/controller"
	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/modelcatalog"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/task"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. Initialize DB (config.db)
	db, dbCleanup, err := datastore.NewSQLiteDB(context.Background(), "config.db")
	if err != nil {
		log.Fatalf("failed to initialize config database: %v", err)
	}
	defer dbCleanup()

	// 1.1 Initialize task DB (task.db)
	taskDB, taskDBCleanup, err := datastore.NewSQLiteDB(context.Background(), "task.db")
	if err != nil {
		log.Fatalf("failed to initialize task database: %v", err)
	}
	defer taskDBCleanup()

	// 2. Run Migrations
	if err := config.Migrate(context.Background(), db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	// 3. Setup Task Manager
	logger, wailsLogH := telemetry.ProvideLogger()
	if err := task.Migrate(context.Background(), taskDB); err != nil {
		log.Fatalf("failed to run task database migrations: %v", err)
	}
	taskStore := task.NewStore(taskDB)
	taskManager := task.NewManager(context.TODO(), logger, taskStore) // Context will be set in startup
	personaProgressNotifier := progress.NewWailsNotifier(logger)
	personaProgressNotifier.SetEventName("persona:progress")

	// 4. Setup Dictionary System (dictionary.db)
	dictDB, dictDBCleanup, err := datastore.NewSQLiteDB(context.Background(), "dictionary.db")
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

	// 5. Setup Config Controller (UIStateStore Wails binding)
	configStore, err := config.NewSQLiteStore(context.Background(), db, logger)
	if err != nil {
		log.Fatalf("failed to initialize config store: %v", err)
	}
	configController := controller.NewConfigController(configStore, logger)
	llmManager := llm.NewLLMManager(logger)
	modelCatalogService := modelcatalog.NewModelCatalogService(configStore, configStore, llmManager, logger)

	// 6. Setup Persona + Parser dependencies for task bridge.
	parserLoader := parser.ProvideParser(configStore)
	llmQueue, err := queue.NewQueue(context.Background(), "llm_queue.db", logger)
	if err != nil {
		log.Fatalf("failed to initialize llm queue: %v", err)
	}
	defer func() {
		_ = llmQueue.Close()
	}()

	queueWorker := queue.NewWorker(llmQueue, llmManager, configStore, configStore, personaProgressNotifier, logger)
	if err := queueWorker.Recover(context.Background()); err != nil {
		log.Printf("failed to recover llm queue worker state: %v", err)
	}

	// Persona data is persisted in dedicated persona.db.
	personaDB, personaDBCleanup, err := datastore.NewSQLiteDB(context.Background(), "persona.db")
	if err != nil {
		log.Fatalf("failed to initialize persona database: %v", err)
	}
	defer personaDBCleanup()

	personaStore := persona.NewPersonaStore(personaDB)
	if err := personaStore.InitSchema(context.Background()); err != nil {
		log.Fatalf("failed to initialize persona schema: %v", err)
	}
	personaGenerator := persona.NewPersonaGenerator(
		persona.NewDefaultDialogueCollector(),
		persona.NewDefaultContextEvaluator(persona.NewDefaultScorer(), persona.NewSimpleTokenEstimator()),
		personaStore,
		configStore,
		configStore,
	)
	personaService := persona.NewService(personaStore, logger)

	// 7. Setup Bridge
	masterPersonaWorkflow := workflow.NewMasterPersonaService(taskManager, logger, parserLoader, personaGenerator, personaProgressNotifier, llmQueue, queueWorker)
	taskManager.RegisterRunner(task.TypePersonaExtraction, masterPersonaWorkflow)
	taskManager.RegisterCompletionHook(task.TypePersonaExtraction, masterPersonaWorkflow.CleanupCompletedTask)
	taskController := controller.NewTaskController(taskManager)
	personaTaskController := controller.NewPersonaTaskController(taskManager, masterPersonaWorkflow)

	// Create an instance of the app structure
	app := NewApp()
	app.SetDictService(dictService)
	app.SetConfigService(configController)

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
			// Wails ログハンドラにランタイムコンテキストを注入（これ以降 emit が有効になる）
			wailsLogH.SetContext(ctx)
			// Pass Wails context to TaskManager and Initialize it
			taskManager.SetContext(ctx)
			configController.SetContext(ctx)
			taskController.SetContext(ctx)
			personaTaskController.SetContext(ctx)
			if err := taskManager.Initialize(ctx); err != nil {
				log.Printf("failed to initialize task manager: %v", err)
			}
			// Dictionary service progress notifier
			wailsNotifier.SetContext(ctx)
			personaProgressNotifier.SetContext(ctx)
		},
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
			taskController,
			personaTaskController,
			configController,
			modelCatalogService,
			personaService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
