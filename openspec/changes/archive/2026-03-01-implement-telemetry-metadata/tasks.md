# Tasks: implement-telemetry-metadata

## 1. Context Extension for Telemetry (context.go)
- [x] 1.1 `pkg/infrastructure/telemetry/context.go` を作成し、カスタムログ属性（`slog.Attr` スライス）を乗せるための専用コンテキストキー（未エクスポート）を定義する。
- [x] 1.2 コンテキストから追加属性スライスを取得する内部関数 `attrsFromContext(ctx) []slog.Attr` を実装する。
- [x] 1.3 `action` および `resource_type` 用に独自の型（`ActionType string`, `ResourceType string`）とシステム横断的な定数群（`ActionImport`, `ResourceDictionary` 等）を定義し、記述の揺れと型安全性を確保する。
- [x] 1.4 指定された `action` やリソース情報をコンテキストの属性として追加・返却する `WithAction(ctx context.Context, action ActionType, resType ResourceType, resID string)` 関数を実装する。

## 2. Provider Extension (provider.go)
- [x] 2.1 `ProvideLogger` を改修し、初期化時に `slog.NewJSONHandler` のベースにグローバル属性（`env`, `app_version`, `service_name`）を自動付与する。
- [x] 2.2 `otelHandler.Handle` メソッドを拡張し、`attrsFromContext(ctx)` で取得した属性をログの `slog.Record` に動的にマージする。

## 3. Performance & Error Utilities
- [x] 3.1 関数実行時間を計測してログに出力するためのユーティリティ `StartSpan(ctx, action ActionType) func()` を作成する。呼び出し側が `defer` で用いることで、終了時に自動で `duration_ms` 属性などが付与されたログを出力させる。
- [x] 3.2 エラーロギングを統一するためのユーティリティ `ErrorAttrs(err error) []slog.Attr` を実装し、`error_code`, `exception_class`, `stack_trace` のキー名でエラー詳細属性を生成できるようにする。

## 4. Verification
- [x] 4.1 `telemetry-metadata/spec.md` に記載された4つの Requirement シナリオに基づき、各機能（グローバルキーの付与、Semantic Actionの付与、パフォーマンス測定、エラーコンテキストの記録）が正しく JSON ログに追加されるか、統合的に動作確認を行う。
