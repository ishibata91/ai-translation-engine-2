package main

import (
	"context"
	"embed"
	"log"

	dictionary_artifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/dictionary_artifact"
	master_persona_artifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/master_persona_artifact"
	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/controller"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
	gatewayconfig "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/configstore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/llmexec"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/modelcatalog"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	dictionary2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/terminology"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
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

	// 1.2 Initialize shared artifact DB (artifact.db)
	artifactDB, artifactDBCleanup, err := datastore.NewSQLiteDB(context.Background(), "artifact.db")
	if err != nil {
		log.Fatalf("failed to initialize artifact database: %v", err)
	}
	defer artifactDBCleanup()

	terminologyDB, terminologyDBCleanup, err := datastore.NewSQLiteDB(context.Background(), "terminology.db")
	if err != nil {
		log.Fatalf("failed to initialize terminology database: %v", err)
	}
	defer terminologyDBCleanup()

	// 2. Run Migrations
	if err := configstore.Migrate(context.Background(), db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}
	if err := translationinput.Migrate(context.Background(), artifactDB); err != nil {
		log.Fatalf("failed to run artifact database migrations: %v", err)
	}
	if err := dictionary_artifact.Migrate(context.Background(), artifactDB); err != nil {
		log.Fatalf("failed to run dictionary artifact migrations: %v", err)
	}
	if err := master_persona_artifact.Migrate(context.Background(), artifactDB); err != nil {
		log.Fatalf("failed to run master persona artifact migrations: %v", err)
	}

	// 3. Setup Task Manager
	logger, wailsLogH := telemetry.ProvideLogger()
	if err := task2.Migrate(context.Background(), taskDB); err != nil {
		log.Fatalf("failed to run task database migrations: %v", err)
	}
	taskStore := task2.NewStore(taskDB)
	taskManager := task2.NewManager(context.TODO(), logger, taskStore) // Context will be set in startup
	personaProgressNotifier := progress.NewWailsNotifier(logger)
	personaProgressNotifier.SetEventName("persona:progress")
	translationFlowProgressNotifier := progress.NewWailsNotifier(logger)
	translationFlowProgressNotifier.SetEventName("translation_flow.terminology.progress")

	// 4. Setup Dictionary System (artifact shared store)
	dictArtifactRepo := dictionary_artifact.NewRepository(artifactDB)
	dictStore := dictionary2.NewDictionaryStore(dictArtifactRepo)
	wailsNotifier := progress.NewWailsNotifier(logger)
	wailsNotifier.SetEventName("dictionary:import_progress") // 必要に応じて変更可能

	// Wails AST needs `New<StructName>` pattern or `&StructName{}` pattern to discover bindings properly
	dictConfig := dictionary2.DefaultConfig()
	dictImporter := dictionary2.NewImporter(dictConfig, dictStore, wailsNotifier, logger)
	dictService := dictionary2.NewDictionaryService(dictStore, dictImporter, logger)

	// 5. Setup Config Controller (UIStateStore Wails binding)
	configStore, err := configstore.NewSQLiteStore(context.Background(), db, logger)
	if err != nil {
		log.Fatalf("failed to initialize config store: %v", err)
	}
	gatewayConfigStore, err := gatewayconfig.NewSQLiteStore(context.Background(), db, logger)
	if err != nil {
		log.Fatalf("failed to initialize gateway config store: %v", err)
	}
	configController := controller.NewConfigController(configStore, logger)
	llmManager := llm.NewLLMManager(logger)
	modelCatalogService := modelcatalog.NewModelCatalogService(configStore, configStore, llmManager, logger)
	modelCatalogController := controller.NewModelCatalogController(modelCatalogService)

	// 6. Setup Persona + Parser dependencies for task bridge.
	parserLoader := skyrim.ProvideParser()
	translationInputRepo := translationinput.NewRepository(artifactDB)
	translationFlowSlice := translationflow.NewService(translationInputRepo)
	llmQueue, err := queue.NewQueue(context.Background(), "llm_queue.db", logger)
	if err != nil {
		log.Fatalf("failed to initialize llm queue: %v", err)
	}
	defer func() {
		_ = llmQueue.Close()
	}()

	queueWorker := queue.NewWorker(llmQueue, llmManager, gatewayConfigStore, gatewayConfigStore, personaProgressNotifier, logger)
	if err := queueWorker.Recover(context.Background()); err != nil {
		log.Printf("failed to recover llm queue worker state: %v", err)
	}
	taskManager.SetTaskRequestCleaner(llmQueue)

	termConfig := &terminology.TermRecordConfig{
		TargetRecordTypes: []string{
			"NPC_:FULL", "NPC_:SHRT",
			"WEAP:FULL", "ARMO:FULL", "AMMO:FULL", "MISC:FULL", "KEYM:FULL", "ALCH:FULL", "BOOK:FULL", "INGR:FULL",
			"SPEL:FULL", "MGEF:FULL", "ENCH:FULL",
			"LCTN:FULL", "CELL:FULL", "WRLD:FULL",
			"MESG:FULL", "QUST:FULL",
		},
	}
	termBuilder := terminology.NewTermRequestBuilder(termConfig)
	termPromptBuilder, err := terminology.NewTermPromptBuilder("")
	if err != nil {
		log.Fatalf("failed to initialize terminology prompt builder: %v", err)
	}
	termSearchStemmer := terminology.NewSnowballStemmer("english")
	termSearcher := terminology.NewSQLiteTermDictionarySearcher(artifactDB, logger, termSearchStemmer)
	termStore := terminology.NewSQLiteModTermStore(terminologyDB, logger)
	if err := termStore.InitSchema(context.Background()); err != nil {
		log.Fatalf("failed to initialize terminology store schema: %v", err)
	}
	termTranslator := terminology.NewTermTranslator(translationInputRepo, termBuilder, termSearcher, termStore, termPromptBuilder, logger)
	translationFlowWorkflow := workflow.NewTranslationFlowService(
		parserLoader,
		translationFlowSlice,
		termTranslator,
		llmexec.NewSyncExecutor(llmManager),
		translationFlowProgressNotifier,
	)

	personaArtifactRepo := master_persona_artifact.NewRepository(artifactDB)
	personaStore := persona.NewPersonaStore(personaArtifactRepo)
	if err := personaStore.InitSchema(context.Background()); err != nil {
		log.Fatalf("failed to initialize persona artifact store: %v", err)
	}
	personaGenerator := persona.NewPersonaGenerator(
		persona.NewDefaultDialogueCollector(),
		persona.NewDefaultContextEvaluator(persona.NewDefaultScorer(), persona.NewSimpleTokenEstimator()),
		personaStore,
		gatewayConfigStore,
		gatewayConfigStore,
	)
	personaService := persona.NewService(personaStore, logger)
	personaController := controller.NewPersonaController(personaService)

	// 7. Setup Bridge
	masterPersonaWorkflow := workflow.NewMasterPersonaService(taskManager, logger, parserLoader, personaGenerator, personaProgressNotifier, llmQueue, queueWorker)
	taskManager.RegisterRunner(task2.TypePersonaExtraction, masterPersonaWorkflow)
	taskManager.RegisterCompletionHook(task2.TypePersonaExtraction, masterPersonaWorkflow.CleanupCompletedTask)
	taskController := controller.NewTaskController(taskManager)
	taskController.SetTranslationFlowWorkflow(translationFlowWorkflow)
	personaTaskController := controller.NewPersonaTaskController(taskManager, masterPersonaWorkflow)
	dictionaryController := controller.NewDictionaryController(dictService)
	fileDialogController := controller.NewFileDialogController()

	// Create an instance of the app structure
	app := NewApp()

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
			dictionaryController.SetContext(ctx)
			fileDialogController.SetContext(ctx)
			modelCatalogController.SetContext(ctx)
			personaController.SetContext(ctx)
			if err := taskManager.Initialize(ctx); err != nil {
				log.Printf("failed to initialize task manager: %v", err)
			}
			// Dictionary service progress notifier
			wailsNotifier.SetContext(ctx)
			personaProgressNotifier.SetContext(ctx)
			translationFlowProgressNotifier.SetContext(ctx)
		},
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			taskController,
			personaTaskController,
			configController,
			dictionaryController,
			fileDialogController,
			modelCatalogController,
			personaController,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
