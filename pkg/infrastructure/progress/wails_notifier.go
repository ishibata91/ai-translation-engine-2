package progress

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsNotifier は Wails のイベントシステムを使って進捗をフロントエンドに送信する。
type WailsNotifier struct {
	wailsCtx  context.Context
	logger    *slog.Logger
	eventName string
}

// NewWailsNotifier は 新しい WailsNotifier を生成する。
func NewWailsNotifier(logger *slog.Logger) *WailsNotifier {
	return &WailsNotifier{
		logger:    logger.With("component", "WailsNotifier"),
		eventName: "task:updated", // デフォルトのイベント名。フロントエンドの TaskStore に合わせる
	}
}

// SetContext は Wails ランタイムコンテキストをセットする。
func (w *WailsNotifier) SetContext(ctx context.Context) {
	w.wailsCtx = ctx
}

// SetEventName は発行するイベント名を変更する。
func (w *WailsNotifier) SetEventName(name string) {
	w.eventName = name
}

// OnProgress は ProgressEvent をフロントエンドに送信する。
func (w *WailsNotifier) OnProgress(_ context.Context, event ProgressEvent) {
	if w.wailsCtx != nil {
		runtime.EventsEmit(w.wailsCtx, w.eventName, event)
	} else {
		w.logger.Warn("wails context is not set, skipping progress event")
	}
}
