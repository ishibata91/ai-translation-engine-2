# Proposal: implement-telemetry-metadata

## Why

`specs/log-guide.md`（AI解析に最適化されたログ設計ガイド）の取り決めに基づき、現行の構造化ログにAIが追跡・推論を行うための「メタデータ」を付与できるようにするため。
現在、`pkg/infrastructure/telemetry/provider.go` では `trace_id` と `span_id` をコンテキストから抽出・記録するにとどまっており、AIによる自律的なデバッグ（`/structured-log-debug`）を実現するためには、より詳細なコンテキスト情報（環境、セマンティクス、エラー詳細など）をログに落とし込む基盤が必要です。

## What Changes

現在の構造化ログ基盤（`pkg/infrastructure/telemetry`）を拡張し、以下のようなメタデータをログ出力時に統一的に付与または抽出できるユーティリティ群を実装します。

- **実行環境キー**: `env`, `service_name`, `app_version` 等の固定情報付与
- **セマンティクスキー**: `action` (イベント名)、`resource_type` 等の業務的コンテキスト付与
- **パフォーマンス/リソースキー**: `duration_ms` の付与機能
- **エラー解決キー**: `error_code`, `exception_class`, `stack_trace` の形式的な付与機能

## Capabilities

- `telemetry-metadata`: 構造化ログに対してAI解析に必要な追加コンテキスト（環境・実行セマンティクス・エラー情報）を付与するためのユーティリティおよび拡張基盤の提供。

## Impact

- `pkg/infrastructure/telemetry/provider.go`: 初期化時に固定メタデータ（環境・バージョンなど）を付与する仕組みの追加。
- `pkg/infrastructure/telemetry/context.go` (新設想定): コンテキストへの属性付与を簡略化するヘルパーの追加。
- `pkg/infrastructure/telemetry/` パッケージを利用する後続のすべてのスライス実装（本チェンジ完了後に段階的に適用される想定のため、本変更単体での広範な既存コード破壊はない）。
