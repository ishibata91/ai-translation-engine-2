## 1. Foundation & Infrastructure
- [x] 1.1 `*sql.DB` を提供する DB 接続プロバイダーの実装（パス解決、プール設定）
- [x] 1.2 `pkg/config_store/migration.go` の実装（テーブル作成ロジック）
- [x] 1.3 `pkg/config_store/provider.go` の実装（Wire ProviderSet 定義）

## 2. Core Implementation
- [x] 2.1 `pkg/config_store/sqlite_store.go` の実装（ConfigStore, UIStateStore, SecretStore インターフェース）
- [x] 2.2 `pkg/config_store/typed_accessor.go` の実装（型安全アクセサ）

## 3. Features & Logic
- [x] 3.1 変更通知 (Watch) ロジックの実装
- [x] 3.2 `slog` による構造化ログと TraceID 伝播の実装

## 4. Verification
- [x] 4.1 パラメタライズドテストの実装 (`pkg/config_store/test`)
- [x] 4.2 期待される動作（保存、取得、Watch、マイグレーション）の検証
