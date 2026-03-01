package telemetry

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsLogEvent はフロントエンドに送信するログイベントの構造体。
// JSON タグはフロントエンドの TypeScript 型定義と一致させる。
type WailsLogEvent struct {
	ID         string         `json:"id"`
	Level      string         `json:"level"`
	Message    string         `json:"message"`
	Timestamp  string         `json:"timestamp"`
	Attributes map[string]any `json:"attributes"`
}

// wailsLogEventName は Wails イベントシステムで使用するイベント名。
const wailsLogEventName = "telemetry.log"

// wailsCtxHolder は Wails ランタイムコンテキストを派生ハンドラ間で共有するためのホルダ。
type wailsCtxHolder struct {
	mu  sync.RWMutex
	ctx context.Context
}

// wailsHandler は slog.Handler を実装し、ログを Wails イベントとして emit する。
// Wails ランタイムコンテキストの設定前はログをドロップする（startup 前に安全）。
type wailsHandler struct {
	next      slog.Handler
	ctxHolder *wailsCtxHolder
	preAttrs  []slog.Attr
}

// newWailsHandler は新しい wailsHandler を生成する。
func newWailsHandler(next slog.Handler) *wailsHandler {
	return &wailsHandler{
		next:      next,
		ctxHolder: &wailsCtxHolder{},
	}
}

// SetContext は Wails ランタイムコンテキストを注入する。
// app.startup() 完了後に呼び出すこと。
func (h *wailsHandler) SetContext(ctx context.Context) {
	h.ctxHolder.mu.Lock()
	defer h.ctxHolder.mu.Unlock()
	h.ctxHolder.ctx = ctx
}

func (h *wailsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *wailsHandler) Handle(ctx context.Context, r slog.Record) error {
	// まず下流ハンドラー（コンソール出力）に渡す
	if err := h.next.Handle(ctx, r); err != nil {
		return err
	}

	// Wails コンテキストが未設定の場合はスキップ
	h.ctxHolder.mu.RLock()
	wCtx := h.ctxHolder.ctx
	h.ctxHolder.mu.RUnlock()

	if wCtx == nil {
		return nil
	}

	// ログイベントを組み立てる
	attrs := make(map[string]any)

	// handler に事前に付与された属性（グローバル属性等）を収集
	for _, a := range h.preAttrs {
		attrs[a.Key] = resolveAttrValue(a.Value)
	}

	// コンテキスト由来の属性を収集
	for _, a := range attrsFromContext(ctx) {
		attrs[a.Key] = resolveAttrValue(a.Value)
	}

	// レコード内の属性を収集
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = resolveAttrValue(a.Value)
		return true
	})

	// UUID 風の ID を生成（タイムスタンプ + ランダム部分は trace_id があれば使用）
	id := generateEventID(attrs)

	event := WailsLogEvent{
		ID:         id,
		Level:      r.Level.String(),
		Message:    r.Message,
		Timestamp:  r.Time.Format(time.RFC3339Nano),
		Attributes: attrs,
	}

	// goroutine で emit して非同期化（バックエンド処理をブロックしない）
	go func() {
		runtime.EventsEmit(wCtx, wailsLogEventName, event)
	}()

	return nil
}

func (h *wailsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newH := &wailsHandler{
		next:      h.next.WithAttrs(attrs),
		ctxHolder: h.ctxHolder,
		preAttrs:  append(append([]slog.Attr{}, h.preAttrs...), attrs...),
	}
	return newH
}

func (h *wailsHandler) WithGroup(name string) slog.Handler {
	return &wailsHandler{
		next:      h.next.WithGroup(name),
		ctxHolder: h.ctxHolder,
		preAttrs:  h.preAttrs,
	}
}

// resolveAttrValue は slog.Value を JSON シリアライズ可能な Go 値に変換する。
func resolveAttrValue(v slog.Value) any {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindGroup:
		m := make(map[string]any)
		for _, a := range v.Group() {
			m[a.Key] = resolveAttrValue(a.Value)
		}
		return m
	case slog.KindAny:
		anyVal := v.Any()
		// JSON マーシャル可能かテスト
		if b, err := json.Marshal(anyVal); err == nil {
			_ = b
		}
		return anyVal
	default:
		return v.String()
	}
}

// generateEventID はイベントの一意 ID を生成する。
// trace_id が存在すればその先頭を使用し、なければタイムスタンプを使う。
func generateEventID(attrs map[string]any) string {
	if tid, ok := attrs["trace_id"].(string); ok && len(tid) >= 8 {
		return tid[:8] + "-" + time.Now().Format("150405.000")
	}
	return time.Now().Format("20060102-150405.000000000")
}
