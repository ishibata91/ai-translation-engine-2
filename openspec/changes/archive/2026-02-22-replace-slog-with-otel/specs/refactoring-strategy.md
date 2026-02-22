# Delta Spec: refactoring-strategy

## Added Requirements

### Requirement: Automatic Trace Correlation
`pkg/` 以下の各コンポーネントで出力される全ての `slog` ログは、OpenTelemetry のトレース情報（TraceID, SpanID）を保持していなければならない。

### Requirement: Context Propagation
全ての `Contract` メソッドおよびその内部処理において、`context.Context` を正しく伝搬させ、`slog.InfoContext` 等の `Context` 付きメソッドを使用しなければならない。

### Requirement: slog-otel Integration
`github.com/samber/slog-otel` を利用し、`slog.Handler` レイヤーで `context.Context` からトレース情報を抽出し、JSON ログの `trace_id` および `span_id` フィールドに自動的に記録されるように設定すること。

## Scenarios

#### Scenario: Log with TraceID
- **WHEN** OTel トレースが開始された context を用いてログを出力する
- **THEN** 出力された JSON ログに、正しい `trace_id` と `span_id` が含まれている

#### Scenario: Log without TraceID
- **WHEN** OTel トレース情報がない context を用いてログを出力する
- **THEN** ログは正常に出力され、`trace_id` 等のフィールドは空（または省略）となるが、エラーにはならない
