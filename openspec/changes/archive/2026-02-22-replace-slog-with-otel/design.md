# Design: slog-otel Integration

## Context
現在、`pkg/` 以下の各スライスでは `*slog.Logger` が DI で注入されて使用されていますが、`context.Context` に含まれるトレース情報がログに出力されていません。`github.com/samber/slog-otel` を導入することで、`slog` のハンドラレベルでトレースIDを抽出し、ログレコードに付与します。

## Goals / Non-Goals
**Goals:**
- `pkg/` 以下の全てのログに `trace_id` と `span_id` を付与する。
- 既存の `slog` 利用箇所を最小限の変更で OTel 対応にする。
- `context.Context` の伝播を徹底する。

**Non-Goals:**
- 独自のロギングライブラリの作成。
- OTel Collector への直接送信の実装（現在はログ出力を介した紐付けに集中）。

## Implementation Details

### 1. Handler Configuration
アプリケーションの初期化プロセス（`main.go` または `infrastructure` レイヤー）において、`slog.Handler` を以下のようにラップします。

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
otelHandler := slogotel.NewHandler(handler)
logger := slog.New(otelHandler)
slog.SetDefault(logger)
```

### 2. Context Propagation
`pkg/` 以下の各メソッドにおいて、`slog.Info(...)` などのコンテキストを取らないメソッドを、`slog.InfoContext(ctx, ...)` に置換します。

対象例:
- `pkg/term_translator/translator.go`
- `pkg/dictionary_builder/importer.go`
- `pkg/config_store/sqlite_store.go`

### 3. Dependency Management
`go.mod` に `github.com/samber/slog-otel` を追加します。

## Risks / Trade-offs
- **パフォーマンス**: 各ログ出力ごとにコンテキストからの抽出処理が走るため、極微量のオーバーヘッドが発生しますが、通常の翻訳処理（LLM呼び出し等）に比べれば無視できるレベルです。
- **依存性**: コミュニティ製のライブラリ（samber/slog-otel）に依存しますが、`slog.Handler` のラッパーとして非常にシンプルであり、リスクは低いと判断します。
