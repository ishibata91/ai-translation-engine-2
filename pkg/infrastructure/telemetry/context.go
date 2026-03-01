package telemetry

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// ---- 型定義 ----------------------------------------------------------------

// ActionType はアクションを表す型安全なラッパー。
// ログの揺れを防ぐため文字列定数と組み合わせて利用する。
type ActionType string

// ResourceType はリソース種別を表す型安全なラッパー。
type ResourceType string

// ---- アクション定数 --------------------------------------------------------

const (
	ActionImport    ActionType = "import"
	ActionExport    ActionType = "export"
	ActionQuery     ActionType = "query"
	ActionCreate    ActionType = "create"
	ActionUpdate    ActionType = "update"
	ActionDelete    ActionType = "delete"
	ActionValidate  ActionType = "validate"
	ActionTranslate ActionType = "translate"
)

// ---- リソース定数 ----------------------------------------------------------

const (
	ResourceDictionary ResourceType = "dictionary"
	ResourceEntry      ResourceType = "entry"
	ResourceTask       ResourceType = "task"
	ResourceUser       ResourceType = "user"
	ResourceFile       ResourceType = "file"
)

// ---- コンテキストキー -------------------------------------------------------

// contextKey は未エクスポートの型で、context.Value のキーとして使用する。
// string や int などとの衝突を回避するため専用型を用いる。
type contextKey int

const (
	attrsKey contextKey = iota
	_                   // 将来の拡張用
)

// ---- 内部ユーティリティ ----------------------------------------------------

// attrsFromContext はコンテキストに格納されたカスタム slog.Attr スライスを返す。
// 格納されていない場合は nil を返す。
func attrsFromContext(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	v, _ := ctx.Value(attrsKey).([]slog.Attr)
	return v
}

// ---- 公開ユーティリティ ----------------------------------------------------

// WithAttrs は指定した slog.Attr を追加したコンテキストを返す。
// 既存の属性は保持される（append によるコピー）。
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing := attrsFromContext(ctx)
	// 既存スライスを変更しないようコピーしてから追加する
	merged := make([]slog.Attr, len(existing), len(existing)+len(attrs))
	copy(merged, existing)
	merged = append(merged, attrs...)
	return context.WithValue(ctx, attrsKey, merged)
}

// WithAction はアクションおよびリソース情報をコンテキストに付与した新しい
// context.Context を返す。後続の全ログにこれらの属性が引き継がれる。
//
// 例:
//
//	ctx = telemetry.WithAction(ctx, telemetry.ActionImport, telemetry.ResourceDictionary, "src-001")
func WithAction(ctx context.Context, action ActionType, resType ResourceType, resID string) context.Context {
	return WithAttrs(ctx,
		slog.String("action", string(action)),
		slog.String("resource_type", string(resType)),
		slog.String("resource_id", resID),
	)
}

// WithRequestID は新しい UUID を生成し、`request_id` としてコンテキストに付与する。
// Wails バインディングの各メソッド冒頭で一度呼ぶだけで、
// Service / Store 層の全ログに自動で request_id が乗る。
//
// 例:
//
//	ctx := telemetry.WithRequestID(a.ctx)
func WithRequestID(ctx context.Context) context.Context {
	return WithAttrs(ctx, slog.String("request_id", uuid.New().String()))
}
