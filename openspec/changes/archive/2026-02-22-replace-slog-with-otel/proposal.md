# Proposal: Replace slog with slog-otel

## Why
現在、`pkg/` 以下の各スライスで `log/slog` を使用していますが、OpenTelemetry との統合が行われておらず、分散トレース（TraceID / SpanID）がログに紐付いていません。
デバッグの容易性を向上させ、リファクタリング戦略で定義された「構造化ログ基盤」を実現するために、`slog` の出力を OpenTelemetry トレースと自動的に連動させる必要があります。

## What Changes
- `pkg/` 以下の全ての `slog` 利用箇所において、`context.Context` を適切に伝播させるよう修正します。
- `github.com/samber/slog-otel` を導入し、`slog.Handler` で TraceID/SpanID を自動的に抽出・埋め込むように構成します。
- `infrastructure` レイヤーでのロガー初期化処理において、`slog-otel` ハンドラを注入するように変更します。

## Capabilities
- `refactoring-strategy`: トレースIDによる横断追跡を可能にする構造化ログ基盤の具体的実装を完了します。

## Impact
- **影響範囲**: `pkg/` 以下のほぼ全てのパッケージ（`term_translator`, `dictionary_builder`, `config_store` 等）。
- **依存関係**: `github.com/samber/slog-otel` が新しく追加されます。
- **後方互換性**: ログの出力形式が JSON である点は変わりませんが、フィールドに `trace_id` と `span_id` が追加されます。
