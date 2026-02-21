# config-store Specification

## Purpose
TBD - created by archiving change implement-config-store-slice. Update Purpose after archive.
## Requirements
### Requirement: SQLite Persistence
`ConfigStore`, `UIStateStore`, `SecretStore` の全データは SQLite データベースファイルに永続化されなければならない (MUST).

#### Scenario: Verify file creation
- **WHEN** 最初の `Set` 操作が行われた場合
- **THEN** 指定されたパスに SQLite データベースファイルが生成される

### Requirement: DB Connection Injection
`SQLiteStore` は `*sql.DB` を DI 経由で受け取らなければならない (MUST).

#### Scenario: Injection check
- **WHEN** プロバイダーから `SQLiteStore` を生成する場合
- **THEN** インフラ層で初期化された `*sql.DB` が渡される

### Requirement: Table Isolation
データは `config`, `ui_state`, `secrets` の3つのテーブルに明確に分離して格納されなければならない (MUST).

#### Scenario: Table validation
- **WHEN** `Migrate` が実行された場合
- **THEN** 3つの独立したテーブルがスキーマに作成される

### Requirement: Type-Safe Access
`config` テーブルの値は、`TypedAccessor` を通じて型安全に取得できなければならない (MUST).

#### Scenario: Get integer configuration with default
- **WHEN** `TypedAccessor.GetInt` を呼び出し、キーが存在しない場合
- **THEN** 引数で指定されたデフォルト値が返される

### Requirement: Change Notification (Watch)
値が変更された際に登録されたコールバックが呼び出されなければならない (MUST).

#### Scenario: Watch configuration change
- **WHEN** `ConfigStore.Watch` で監視中に、`ConfigStore.Set` で値が変更された場合
- **THEN** 新旧の値を含む変更イベントを引数としてコールバックが発火する

### Requirement: Auto Migration
アプリケーション起動時に、必要なテーブル構造が自動的に作成されなければならない (MUST).

#### Scenario: Migration check
- **WHEN** アプリが起動した場合
- **THEN** `schema_version` テーブルが作成され、最新バージョンまで更新される

### Requirement: JSON Support
`UIStateStore` は構造化データを保存できなければならない (MUST).

#### Scenario: Store JSON in UIStateStore
- **WHEN** `UIStateStore.SetJSON` でデータを保存し、`GetJSON` で取得した場合
- **THEN** データが正しく元の構造体に戻る

