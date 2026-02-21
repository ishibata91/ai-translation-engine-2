## Context
`pkg/config_store` には設定保存のためのインターフェース（Contract）が定義されているが、具体的な永続化ロジック（SQLite実装）が未実装である。
`refactoring_strategy.md` は、各スライスが自律的に自身の状態を管理し、設定を `ConfigStoreSlice` から取得する Vertical Slice Architecture を推奨している。

## Goals / Non-Goals
**Goals:**
- SQLite をバックエンドとした `ConfigStore`, `UIStateStore`, `SecretStore` の具象実装の提供。
- `TypedAccessor` を介した、`config` テーブルへの型安全なアクセス（Int, Bool等）。
- `Watch` 機能による、設定変更時のリアルタイム通知。
- 自動マイグレーション機能による、テーブル構成の維持。
- `refactoring_strategy.md` 準拠の構造化ログと TraceID の伝播。

**Non-Goals:**
- `secrets` テーブルの暗号化（将来の拡張として定義、今回はプレーンテキスト）。
- GUI 設定画面の実装（これは UI 側の責務）。

## Proposed Design
### 1. データベース構造 (SQLite)
3つの主要テーブルを作成する。
- `config`: 名前空間とキーに基づく文字列保存用（JSON不許可）。
- `ui_state`: UI状態（パネルサイズ等）のJSON保存用。
- `secrets`: 機密情報（APIキー等）用。専用テーブルに分離。

### 2. DBインフラストラクチャ管理
`refactoring_strategy.md` に基づき、以下の通りDBインフラを構築する。
- **DIによる外部注入**: `SQLiteStore` は具象的なファイルパスへの依存を避け、`*sql.DB`（接続プール済みのインスタンス）を DI (Wire) 経由で受け取る。
- **データディレクトリの解決**: DBファイルの場所（例: `%APPDATA%/ai-translation-engine-2/config.db`）の解決は Provider 層で行い、`*sql.DB` を生成する。
- **マイグレーションの自律性**: スライス自身が `Migrate(ctx)` メソッドを持ち、DAOの初期化時にテーブルの存在を確認する。

### 3. コンポーネント構成
- **`SQLiteStore`**: 3つのインターフェースを単一の構造体で実装。`sync.RWMutex` でスレッドセーフを確保し、`map[string][]ChangeCallback` でリスナーを管理。
- **`TypedAccessor`**: 既存の `ConfigStore` をラップ。`strconv` を使用して文字列から型への変換、およびエラー時のデフォルト値返却を行う。

### 3. 変更通知 (Watch)
- `Set` メソッドが呼ばれた際、値に変更があれば登録されたコールバックを同期的に実行。

### 4. Wire 提供 (DI)
- `pkg/config_store/provider.go` にて `SQLiteStore`, `TypedAccessor` を提供する。

## Risks / Trade-offs
- **SQLite 選択**: CGO 不要で Windows でも動作が安定している `modernc.org/sqlite` を採用。
- **メモリ内 Watcher**: アプリケーション終了時に消失するため、再起動後は再登録が必要（構成上許容）。
