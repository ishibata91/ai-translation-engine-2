# Proposal: Implement ConfigStoreSlice

## Goal
`pkg/config_store` に `ConfigStoreSlice` の具象実装を導入し、SQLite を用いた永続化、型安全なアクセス、および変更通知機能を実現する。
これにより、アプリケーション設定、UI状態、機密情報の一元管理と永続化が可能になる。

## Why
- 現状、`pkg/config_store` にはインターフェース定義（Contract）のみが存在し、具象実装がない。
- 各スライスやUIが設定を永続化するための標準的な手段が欠如している。
- `refactoring_strategy.md` に基づく Vertical Slice アーキテクチャの一部として、インフラストラクチャ層の確立が必要。

## What Changes
- **SQLite Adapter 実装**: `database/sql` と `modernc.org/sqlite` を使用して、`config`, `ui_state`, `secrets` テーブルを管理する。
- **Contract の実装**: `ConfigStore`, `UIStateStore`, `SecretStore` インターフェースの具象実装。
- **TypedAccessor 実装**: `config` テーブルに対する型安全な（Int, Bool, Float等）アクセスを提供。
- **マイグレーション機能**: 起動時の自動テーブル作成とスキーマバージョン管理。
- **リアクティブ通知**: `Watch` メソッドによる設定変更の即時通知（Go Channel と Callback）。
- **構造化ログ & TraceID**: `refactoring_strategy.md` に準拠した `slog` による Entry/Exit ログと TraceID の伝播。

## Impact
- **影響範囲**: `pkg/config_store`
- **依存関係**: `github.com/mattn/go-sqlite3`, `github.com/google/wire`
- **後続作業**: 他のスライス（LLM, UI等）での設定利用が可能になる。

## Capabilities
### New Capabilities
- `config-store`: アプリケーション全体の設定・状態・機密情報の永続化と通知を提供する共通インフラ。

### Modified Capabilities
<!-- なし -->
