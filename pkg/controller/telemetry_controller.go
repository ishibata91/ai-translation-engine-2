package controller

import (
	"context"
	"log/slog"
	"time"
)

// FrontendLogRequest はフロントエンドから受け取るログリクエスト。
// JSON タグはフロントエンドの TypeScript 型定義と一致させる。
type FrontendLogRequest struct {
	// Level はログレベル文字列: "DEBUG" | "INFO" | "WARN" | "ERROR"
	Level string `json:"level"`
	// Message はログメッセージ本文。
	Message string `json:"message"`
	// Timestamp はフロントエンドで生成した ISO 8601 タイムスタンプ（省略可）。
	// 省略された場合はバックエンド受信時刻を使用する。
	Timestamp string `json:"timestamp,omitempty"`
	// Attrs は任意の追加属性。
	Attrs map[string]string `json:"attrs,omitempty"`
}

// TelemetryController はフロントエンド発のログをバックエンドに橋渡しする Wails バインディング。
// フロントエンドで呼んだログは slog 経由でコンソール・ファイル（logs/YYYY-MM-DD.jsonl）に書き込まれる。
type TelemetryController struct {
	ctx    context.Context
	logger *slog.Logger
}

// NewTelemetryController はコントローラーを生成する。
func NewTelemetryController(logger *slog.Logger) *TelemetryController {
	if logger == nil {
		logger = slog.Default()
	}
	return &TelemetryController{
		ctx:    context.Background(),
		logger: logger.With("source", "frontend"),
	}
}

// SetContext は Wails ランタイムコンテキストを注入する。
// app.startup() 後に呼ぶこと。
func (c *TelemetryController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// WriteLog はフロントエンドからのログエントリをバックエンドに記録する。
// 記録先: stdout（JSON）+ logs/YYYY-MM-DD.jsonl
func (c *TelemetryController) WriteLog(req FrontendLogRequest) error {
	// タイムスタンプをパース（フロントが渡した時刻を使う）
	var t time.Time
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, req.Timestamp); err == nil {
			t = parsed
		}
	}
	if t.IsZero() {
		t = time.Now()
	}

	// ログレベルを解決
	level := parseLevel(req.Level)

	// slog.Record を組み立てて既存ハンドラーチェーン（console + file）に流す
	r := slog.NewRecord(t, level, req.Message, 0)
	for k, v := range req.Attrs {
		r.AddAttrs(slog.String(k, v))
	}

	// logger のハンドラー経由で出力（file_handler が受け取り logs/ に書き込む）
	return c.logger.Handler().Handle(c.ctx, r)
}

// parseLevel は文字列をslog.Levelに変換する。未知の場合は INFO を返す。
func parseLevel(s string) slog.Level {
	switch s {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
