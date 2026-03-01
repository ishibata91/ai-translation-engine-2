# 差分仕様書: sync-dictionary-backend

この差分仕様書は、新しいデータベーススキーマ（`dictionary.db`）および Dictionary Builder のUI要件に合わせて `dictionary` スライスを更新します。

## 更新された要件

### 要件: 最新DBスキーマの準拠 (dlc_sources / dlc_dictionary_entries)
`DictionaryStore` は、永続化レイヤーとして `dictionary.db` を使用し、`specs/database_erd.md` で定義されたテーブルを実装しなければならない。

#### シナリオ: スキーマ移行
- **WHEN** スライスが初期化されるとき
- **THEN** `dictionary.db` 内に `dlc_sources` および `dlc_dictionary_entries` テーブルが存在することを確認する。
- **AND** レガシーな `dictionary_entries` テーブルは無視するか、安全に削除すること。

### 要件: 辞書ソースの CRUD
バックエンドは辞書ソースを管理するためのメソッドを提供しなければならない。

#### シナリオ: ソースの一覧表示
- **WHEN** `GetSources` が呼び出されたとき
- **THEN** メタデータ（id, file_name, status, entry_count 等）を含む `dlc_sources` の全レコードを返さなければならない。

#### シナリオ: ソースの削除
- **WHEN** `DeleteSource(id)` が呼び出されたとき
- **THEN** `dlc_sources` からそのソースレコードを削除しなければならない。
- **AND** `dlc_dictionary_entries` に関連付けられているすべてのエントリをカスケード削除しなければならない。

### 要件: 辞書エントリの CRUD (GridEditor サポート)
UI でのインライン編集を可能にするため、バックエンドは個別のエントリ操作をサポートしなければならない。

#### シナリオ: ソースに紐づくエントリの取得
- **WHEN** `GetEntriesBySourceID(sourceID)` が呼び出されたとき
- **THEN** そのソースに関連付けられたすべての `dlc_dictionary_entries` を返さなければならない。

#### シナリオ: エントリの更新
- **WHEN** `UpdateEntry(term)` が呼び出されたとき
- **THEN** `dlc_dictionary_entries` 内の特定の ID に対して `source_text` または `dest_text` を更新しなければならない。

#### シナリオ: エントリの削除
- **WHEN** `DeleteEntry(id)` が呼び出されたとき
- **THEN** `dlc_dictionary_entries` から特定のエントリを削除しなければならない。

### 要件: Wails サービスバインディング (DictionaryService)
内部のスライスロジックを Wails フロントエンドに橋渡しするための新しい `DictionaryService` を実装しなければならない。

#### シナリオ: フロントエンド連携
- **WHEN** Wails アプリが起動するとき
- **THEN** `DictionaryService` がバインディング用として登録されていなければならない。
- **AND** 上記で定義されたすべての CRUD メソッドが、タスク/UI レイヤーからアクセス可能でなければならない。

### 要件: 進捗報告とメタデータを伴うインポート処理
`DictionaryImporter` は、メタデータを登録し、進捗を通知するよう調整されなければならない。

#### シナリオ: ファイルのインポート
- **WHEN** ファイルインポートが開始されたとき
- **THEN** `dlc_sources` レコードを `status: "IMPORTING"` で作成しなければならない。
- **AND** XML のトークンがパースされ、バッチ単位で保存される際、`pkg/infrastructure/progress`（または同等の通知機構）を介して進捗状況を送信しなければならない。
- **AND** 完了時に、`status` は `"COMPLETED"` になり、`entry_count` は実際にインポートされたレコード数に更新されなければならない。

## 既存仕様書への影響
- `specs/dictionary/spec.md` のセクション 3（カプセル化された永続化）を、2テーブル形式のスキーマを使用するように上書き。
- `specs/dictionary/spec.md` の要件を、ソース管理およびエントリ CRUD を含むように拡張。
