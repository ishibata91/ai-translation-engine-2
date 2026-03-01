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
	host, _ := os.Hostname()
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

// ProvideLogger returns a *slog.Logger configured with slog-otel handler.
// グローバル属性（env, app_version, service_name, host_name）がすべての
// ログに自動付与される。
func ProvideLogger() *slog.Logger {
	ga := globalAttrsFromEnv()

	// Base JSON handler（グローバル属性をベース属性として付与）
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{
		slog.String("env", ga.Env),
		slog.String("app_version", ga.AppVersion),
		slog.String("service_name", ga.ServiceName),
		slog.String("host_name", ga.HostName),
	})

	// Wrap with a custom handler to include trace/span IDs and custom context attrs
	otelH := &otelHandler{next: baseHandler}

	logger := slog.New(otelH)

	// Set as default to catch any direct slog calls
	slog.SetDefault(logger)

	return logger
}

// ProviderSet provides the logger for dependency injection.
var ProviderSet = wire.NewSet(ProvideLogger)

type otelHandler struct {
	next slog.Handler
}

func (h *otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle はログレコードを処理する。
// attrsFromContext でカスタム属性（request_id, action, resource_type 等）を
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
