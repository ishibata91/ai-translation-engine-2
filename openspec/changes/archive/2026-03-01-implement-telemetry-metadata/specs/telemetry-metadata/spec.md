# telemetry-metadata

## ADDED Requirements

### Requirement: Global Telemetry Context (執行環境キーの付与)
アプリケーション起動時または提供される Logger の初期化時において、すべてのログに一意または固定で付与されるべき属性（`env`, `app_version`, `service_name`, `host_name` 等）を `slog.Handler` で自動的に注入する機能を備えなければならない。

#### Scenario: Logger Output with Global Context
- **WHEN** telemetry パッケージから提供されるデフォルトの Logger を使用してログを出力する
- **THEN** 出力された JSON ログには `env`, `app_version`, `service_name` などの固定キーが含まれている

### Requirement: Semantic Actions (セマンティクスキーの付与)
特定の一連の処理（イベントやアクション）を開始する際、`context.Context` に対して `action`, `resource_type`, `resource_id` などのセマンティクス情報を追加できる機能を提供しなければならない。これにより、後続の処理で出力される全てのログにそのセマンティクス情報が引き継がれる。

#### Scenario: Execute Action with Semantics
- **WHEN** `telemetry.WithAction(ctx, "ImportDictionary", "Source", "123")` のようにコンテキストにセマンティクスを付与し、そのコンテキストでログ出力を行う
- **THEN** 以降のその `context` に紐づく全てのログには `action="ImportDictionary"`, `resource_type="Source"`, `resource_id="123"` が付与される

### Requirement: Performance Tracking (パフォーマンス測定)
関数の実行開始から終了までの時間を計測し、`duration_ms` としてログに記録するためのユーティリティ（例: `telemetry.StartSpan(ctx, ...)` とその `defer span.End()`）を提供しなければならない。

#### Scenario: End of Span Logging
- **WHEN** 処理の終了時に `span.End()` などのユーティリティ関数が呼ばれる
- **THEN** 自動的に終了時のログ（Exit log）が出力され、`duration_ms`（ミリ秒単位の処理時間）が含まれる

### Requirement: Error Context (エラー解決キーへの対応)
エラー発生時に、`error_code`, `exception_class`, `stack_trace` といったエラーの詳細情報を統一されたキー名でログに記録するためのヘルパーまたは属性生成機能（例: `telemetry.ErrorAttrs(err)`）を提供しなければならない。

#### Scenario: Logging an Error
- **WHEN** エラーが発生し、エラーオブジェクトと共に `telemetry.ErrorAttrs(err)` を用いてログを出力する
- **THEN** 出力された JSON ログには `error_code`, `exception_class`, `stack_trace` といった構造化されたエラー情報が含まれる
