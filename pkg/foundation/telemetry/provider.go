package telemetry

import (
	"context"
	"log/slog"
	"os"
	"runtime"

	"github.com/google/wire"
)

// GlobalAttrs はアプリ起動時に確定する固定メタデータを保持する。
// ProvideLogger が呼ばれた時点で一度だけ評価される。
type GlobalAttrs struct {
	Env         string
	AppVersion  string
	ServiceName string
	HostName    string
}

// globalAttrsFromEnv は環境変数から GlobalAttrs を生成する。
// 未設定の場合はデフォルト値を使用する。
func globalAttrsFromEnv() GlobalAttrs {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "unknown"
	}
	service := os.Getenv("SERVICE_NAME")
	if service == "" {
		service = "ai-translation-engine"
	}
	host, err := os.Hostname()
	if err != nil {
		host = ""
	}
	if host == "" {
		host = runtime.GOOS
	}
	return GlobalAttrs{
		Env:         env,
		AppVersion:  version,
		ServiceName: service,
		HostName:    host,
	}
}

// LoggerSet はアプリ起動時に生成されるロガーとその関連ハンドラーのセット。
type LoggerSet struct {
	Logger *slog.Logger
	WailsH *WailsHandler
	FileH  *fileHandler
	LogDir string
}

// ProvideLogger returns a *slog.Logger configured with broadcast handler.
// グローバル属性（env, app_version, service_name, host_name）がすべての
// ログに自動付与される。
//
// 出力先:
//   - stdout (JSON)
//   - {cwd}/logs/YYYY-MM-DD.jsonl (ファイル)
//   - Wails イベント（startup 後）
//
// 戻り値の *WailsHandler は app.startup() 後に SetContext を呼んで Wails context を注入すること。
// 戻り値の *fileHandler は app.shutdown() 後に Close を呼ぶこと。
func ProvideLogger() (*slog.Logger, *WailsHandler, *fileHandler) {
	ga := globalAttrsFromEnv()

	globalSlogAttrs := []slog.Attr{
		slog.String("env", ga.Env),
		slog.String("app_version", ga.AppVersion),
		slog.String("service_name", ga.ServiceName),
		slog.String("host_name", ga.HostName),
	}

	// Base JSON handler（コンソール出力）
	consoleHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}).WithAttrs(globalSlogAttrs)

	// otelHandler でコンテキスト属性（trace_id 等）を自動付与
	otelH := &otelHandler{next: consoleHandler}

	// ファイルハンドラー（{cwd}/logs/YYYY-MM-DD.jsonl）
	logDirPath := resolveLogDir()
	fileH, err := newFileHandler(logDirPath, otelH)
	if err != nil {
		// ファイルハンドラー生成失敗はコンソールのみにフォールバック
		_ = err
		fileH = &fileHandler{baseDir: logDirPath, next: otelH}
	}
	fileHWithAttrs := fileH.WithAttrs(globalSlogAttrs).(*fileHandler)

	// Wails broadcast handler（startup 前は emit をスキップ）
	wH := newWailsHandler(fileHWithAttrs)
	// グローバル属性を wailsHandler にも付与しておく（フロントエンドに届ける）
	wHWithAttrs := wH.WithAttrs(globalSlogAttrs).(*wailsHandler)

	logger := slog.New(wHWithAttrs)

	// Set as default to catch any direct slog calls
	slog.SetDefault(logger)

	return logger, wH, fileHWithAttrs
}

// StartCleanerWithContext はバックグラウンドのログ掃除ゴルーチンを起動する。
// OnStartup から呼ぶことで ctx キャンセル（アプリ終了）と連動して停止できる。
func StartCleanerWithContext(ctx context.Context, fh *fileHandler) {
	StartLogCleaner(ctx, fh.baseDir, LogRetentionDays)
}

// provideWireLogger exposes only the logger for Wire provider signatures.
func provideWireLogger() *slog.Logger {
	logger, _, _ := ProvideLogger()
	return logger
}

// ProviderSet provides the logger for dependency injection.
var ProviderSet = wire.NewSet(provideWireLogger)

type otelHandler struct {
	next slog.Handler
}

func (h *otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle はログレコードを処理する。
// attrsFromContext でカスタム属性（trace_id, action, resource_type 等）を
// コンテキストから取り出してレコードに付与する。
func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	// Custom context attributes (request_id, action, resource_type, resource_id, etc.)
	customAttrs := attrsFromContext(ctx)
	if len(customAttrs) > 0 {
		r.AddAttrs(customAttrs...)
	}
	return h.next.Handle(ctx, r)
}

func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{next: h.next.WithAttrs(attrs)}
}

func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{next: h.next.WithGroup(name)}
}
