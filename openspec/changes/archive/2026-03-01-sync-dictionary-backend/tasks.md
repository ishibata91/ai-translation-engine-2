# タスク: sync-dictionary-backend

新しい UI 要件および ERD スキーマに合わせた、辞書バックエンド同期の実装。

## 1. スキーマとモデル

- [x] 1.1 `pkg/dictionary/dto.go` を `specs/database_erd.md` に合わせた `DictSource` および `DictTerm` 構造体に更新
- [x] 1.2 `pkg/dictionary/contract.go` に `dlc_sources` および `dlc_dictionary_entries` の CRUD 操作を追加
- [x] 1.3 `specs/dictionary/spec.md` を新しいバックエンド要件で更新（メインの仕様書を確定）
- [x] 1.4 `specs/dictionary/dictionary_class_diagram.md` を新しいインターフェースとクラスで更新

## 2. Store の実装 (`pkg/dictionary/store.go`)

- [x] 2.1 `NewDictionaryStore` または初期化ロジックを `dictionary.db` を使用するように更新
- [x] 2.2 `dlc_sources` の CRUD（`GetSources`, `CreateSource`, `UpdateSourceStatus`, `DeleteSource`）を実装（カスケード削除を含む）
- [x] 2.3 `SaveTerms` を `dlc_dictionary_entries` テーブルを使用し、`source_id` にリンクするように更新
- [x] 2.4 `dlc_dictionary_entries` の CRUD（`GetEntriesBySourceID`, `UpdateEntry`, `DeleteEntry`）を実装

## 3. Importer ロジック (`pkg/dictionary/importer.go`)

- [x] 3.1 `ImportXML` を `dlc_sources` のライフサイクル（PENDING -> IMPORTING -> COMPLETED/ERROR）を扱うように更新
- [x] 3.2 `pkg/infrastructure/progress`（または同等の通知システム）と統合し、バッチ処理中に進捗をエミットするように修正
- [x] 3.3 `ImportXML` が DB 内でメタデータ（fileName, fileSize 等）を正確に記録するように確保

## 4. サービス & API レイヤー (`pkg/dictionary/service.go`)

- [x] 4.1 UI レベルのアクションをオーケストレートする `DictionaryService` を作成
- [x] 4.2 `DictionaryService` メソッドを `DictionaryStore` および `DictionaryImporter` の呼び出しにマッピング
- [x] 4.3 `main.go` またはアプリケーションのバインディング設定で `DictionaryService` を Wails バインディング用に登録

## 5. 検証とテスト

- [x] 5.1 `pkg/dictionary/importer_test.go` を新しいスキーマおよびマルチソースアーキテクチャに合わせて更新または書き換え
- [x] 5.2 Wails バインディングが正しく生成され、Wails 開発環境からアクセス可能であることを確認
- [x] 5.3 `DictionaryBuilder.tsx` ページの機能（一覧、インポート、編集）を手動で検証
