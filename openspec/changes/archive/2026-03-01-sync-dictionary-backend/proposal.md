# Proposal: sync-dictionary-backend

## Why

UI側の要件（Dictionary Builder画面、データテーブル、詳細ペインなど）と最新のER図（`dlc_sources`, `dlc_dictionary_entries`）に合わせるため、既存のバックエンド（`pkg/dictionary`）の連携部分に差異と不足が生じています。具体的には、旧スキーマ（`dictionary_entries`）に依存している永続化処理の修正、インポートソースの一覧取得・削除機能の実装、辞書エントリのCRUD（GridEditor向け）、およびWailsのフロントエンドからこれらを呼び出すためのサービスバインディングが欠けているため、これらを解消するべくバックエンドの同期を行います。

## What Changes

1. **DBスキーマ・DTOの同期**: `dictionary_entries` から `dlc_sources` と `dlc_dictionary_entries` のテーブル構成へ移行し、`DictTerm` DTOをER図に合わせて修正。
2. **辞書ソース管理APIの追加**: `contract.go`, `store.go` に辞書ソース（`dlc_sources`）の一覧取得および削除用メソッドを追加。
3. **辞書エントリのUI向けAPI追加**: 辞書ソースIDに紐づくエントリ一覧の取得、エントリの部分更新・削除メソッドを追加し、GridEditorのインライン編集に対応。
4. **Wailsサービスバインディング**: `DictionaryService` を作成し、フロントエンドから `GetSources`, `GetEntries`, `ImportXML` 等を呼び出せるようにする。
5. **Importerの仕様変更**: インポート処理でファイルのメタデータを `dlc_sources` に記録してからパースを実行し、進捗は既存タスク管理インフラを介して通知するように変更。
6. **仕様書・クラス図の更新**: `spec.md` および `dictionary_class_diagram.md` を最新のER図と同期。

## Capabilities

- `update-dictionary-schema`: DBスキーマとDTOを最新ER図 (`dlc_sources`, `dlc_dictionary_entries`) に同期する機能
- `dictionary-source-management`: 辞書ソースの一覧取得および削除機能
- `dictionary-entry-management`: 辞書エントリの一覧取得、更新（インライン編集）、削除機能
- `wails-dictionary-binding`: 辞書管理APIをフロントエンドから呼び出すためのWailsバインディング
- `import-progress-tracking`: ジョブマネージャ等のインフラにインポート進捗を通知する機能の実装

## Impact

- `pkg/dictionary/contract.go` (インターフェースの大幅変更)
- `pkg/dictionary/store.go` (SQLの修正とCRUDメソッドの追加)
- `pkg/dictionary/dto.go` (構造体の定義変更)
- `pkg/dictionary/importer.go` (メタデータのDB登録、進捗通知の実装)
- Wails向けバインディング (`pkg/dictionary/service.go` のようなファイルが新規追加される可能性)
- `specs/dictionary/spec.md` (仕様書の更新)
- `specs/dictionary/dictionary_class_diagram.md` (クラス図の更新)
