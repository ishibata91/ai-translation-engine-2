# Design: implement-telemetry-metadata

## Context
`specs/log-guide.md` に基づいたAI向けの構造化ログを出力するため、既存の `pkg/infrastructure/telemetry` に対してメタデータ（実行環境情報、セマンティック情報、エラー詳細など）を容易に付与・抽出できる拡張を行う。現在は `slog-otel` を用いて TraceID および SpanID のみをコンテキストから抽出してログに出力しているが、これに加え各種業務キーを出力に含める仕組みが必要である。

## Goals / Non-Goals
**Goals:**
- `telemetry` パッケージを利用して初期化された `slog.Logger` に対し、`env`, `app_version` 等の固定情報を自動付与する仕組みを実装する。
- リクエスト/イベント固有のセマンティクス (`action`, `resource_type`, `resource_id`) を `context.Context` を介してログ伝播させるユーティリティを提供する。
- `duration_ms` (実行時間) の計測とログ出力を自動化するユーティリティを提供する。
- エラーハンドリング時に `error_code`, `exception_class`, `stack_trace` 属性を容易に生成するユーティリティを提供する。

**Non-Goals:**
- 全ての既存スライスへのこの実装の即座の適用レイアウト（本チェンジ完了後、各機能開発の中で順次適用するものとする）。
- OpenTelemetry によるトレース（TraceID / SpanID）の基盤自体の変更。

## Architecture / Implementation Plan

1. **Global Context の追加 (Provider 拡張)**
   `provider.go`（`ProvideLogger`）を拡張し、グローバルで付与すべき属性（`env`, `app_version`, `service_name`, など）を環境変数等から取得して `slog.NewJSONHandler` のベース属性または `otelHandler` で追加するように変更する。

2. **Context ベースの属性伝播 (context.go 再設計)**
   `context.Context` 内にカスタムのログ属性（`slog.Attr` スライス）を乗せるための専用キーと、抽出・追加用関数（`WithLogAttrs` 等）を作成する。
   `otelHandler.Handle` メソッド内で、コンテキストにセットされているこれらの属性も取得してログレコードに埋め込むようにする。これにより、コンテキストを引き回すだけでログに情報が引き継がれる。

3. **Semantic Action 用ヘルパーの作成**
   `WithAction(ctx, action, resourceType, resourceID)` ヘルパーを提供し、上記 Context 伝播の仕組みを使ってセマンティクスを持った新しい `context.Context` を返す。
   ログの解釈揺れを防ぐため、引数となる `action` (例: `ActionImport`, `ActionQuery`) や `resourceType` (例: `ResourceDictionary`, `ResourceUser`) は **独自の型（Type）および定数群（Constants）として定義** し、型安全性を確保する。

4. **Performance Measurement 実装**
   `StartSpan(ctx, action)` 関数の実装。内部で開始時刻を記録し、終了時にログ出力と `duration_ms` を付与するコールバック関数を返す構成とする。
   呼び出し側は `defer telemetry.StartSpan(ctx, "Import")()` のように簡潔に記述できる。

5. **Error Attributes ヘルパー**
   `ErrorAttrs(err error)` の実装。`error` インターフェースを調べ、標準の `stack_trace` や型の名前（`exception_class`）を `slog.Group` あるいは一連の `slog.Attr` として返す。

## Decisions

- **属性伝播方式の選択**: `slog-otel` が行っているのと同様に、コンテキスト経由で暗黙的に属性を伝搬させる設計（`otelHandler` でログ出力時に `context` にアクセスして追加属性を結合するアプローチ）を採用する。これにより、各関数側で毎回 `slog.With()` を呼び出してロガーを持ち回る必要がなくなり、純粋な依存性排除（パッケージへの依存度低下）が期待できる。
- **パフォーマンス計測関数の形式**: Goの慣習に従い、`defer func()()` パターンを利用しつつ、スコープ内のロガーにも影響を与えられるようにする。

## Risks / Trade-offs

- **コンテキストのオーバーヘッド**: 全てのログ処理時に `context.Value` から属性をフェッチするため、わずかな実行時オーバーヘッドが生じる。
- **型不整合のリスク**: コンテキスト内の値は `interface{}` として扱われるため、パッケージ内部で型安全なラッパー（未エクスポート型のコンテキストキー）を確実に利用する必要がある。
