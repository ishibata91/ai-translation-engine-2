# デザイン: sync-dictionary-backend

## コンテキスト

フロントエンドの辞書構築UI（Dictionary Builder）の実装が進むにつれ、現在の `pkg/dictionary` の実装が最新のER図およびUI要件（ソース一覧、エントリ編集、進捗表示）と乖離していることが判明した。本デザインでは、これらを完全に同期させ、Wails経由でセキュアかつ効率的に操作するためのバックエンド設計を定義する。

## 決定事項

- **DBテーブルの刷新**: 旧 `dictionary_entries` を廃止し、ER図に基づき `dlc_sources` (ソースメタデータ) と `dlc_dictionary_entries` (個別単語データ) の2テーブル構成に移行する。保存先DBは `dictionary.db` とする。
- **DictionaryService の導入**: `DictionaryImporter` や `DictionaryStore` を直接Wailsにさらすのではなく、UI向けのユースケースをまとめた `DictionaryService` を作成し、これをWailsのバインディング対象とする。
- **進捗通知の統合**: インポート中の進捗は、`pkg/infrastructure/progress` パッケージ（または同等の通知機構）を使用してフロントエンドのジョブマネージャにリアルタイムで反映させる。
- **エントリ編集のサポート**: `SaveTerms` による一括登録だけでなく、GridEditorからの個別更新 (`UpdateEntry`) および削除 (`DeleteEntry`) をサポートする。

## 実装の詳細

### 1. DTOの定義変更 (`pkg/dictionary/dto.go`)
ER図に合わせてフィールドを修正・追加する。

```go
type DictSource struct {
    ID              int64     `json:"id"`
    FileName        string    `json:"file_name"`
    Format          string    `json:"format"`
    FilePath        string    `json:"file_path"`
    FileSize        int64     `json:"file_size_bytes"`
    EntryCount      int       `json:"entry_count"`
    Status          string    `json:"status"` // PENDING, IMPORTING, COMPLETED, ERROR
    ErrorMessage    string    `json:"error_message,omitempty"`
    ImportedAt      *time.Time `json:"imported_at,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
}

type DictTerm struct {
    ID         int64  `json:"id"`
    SourceID   int64  `json:"source_id"`
    EDID       string `json:"edid"`
    RecordType string `json:"record_type"`
    Source     string `json:"source_text"`
    Dest       string `json:"dest_text"`
}
```

### 2. インターフェースの拡張 (`pkg/dictionary/contract.go`)

```go
type DictionaryStore interface {
    // 辞書ソース管理
    GetSources(ctx context.Context) ([]DictSource, error)
    CreateSource(ctx context.Context, src *DictSource) (int64, error)
    UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error
    DeleteSource(ctx context.Context, id int64) error

    // 辞書エントリ管理
    GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]DictTerm, error)
    SaveTerms(ctx context.Context, terms []DictTerm) error // バッチ挿入/アップサート
    UpdateEntry(ctx context.Context, term DictTerm) error
    DeleteEntry(ctx context.Context, id int64) error
}
```

### 3. Wails 連携サービス (`pkg/dictionary/service.go`)
これが Wails から呼び出されるメインの入り口となる。

```go
type DictionaryService struct {
    importer DictionaryImporter
    store    DictionaryStore
}

func (s *DictionaryService) GetSources(ctx context.Context) ([]DictSource, error)
func (s *DictionaryService) DeleteSource(ctx context.Context, id int64) error
func (s *DictionaryService) GetEntries(ctx context.Context, sourceID int64) ([]DictTerm, error)
func (s *DictionaryService) UpdateEntry(ctx context.Context, term DictTerm) error
func (s *DictionaryService) StartImport(ctx context.Context, filePath string) (int64, error) // タスクIDまたはソースIDを返す
```

### 4. インポート処理のフロー
1. ユーザーがファイルを選択 -> `StartImport(path)` 呼び出し。
2. `dlc_sources` に `PENDING` 状態でレコード作成。
3. 非同期（または Wails の Task 経由）で `ImportXML` を実行。
4. `ImportXML` 内で `progress` パッケージを使用して UI に進捗を送信。
5. 完了時に `dlc_sources` のステータスを `COMPLETED` に更新し、`entry_count` をセット。

## リスク / トレードオフ

- **既存テストの破損**: 現在の `importer_test.go` が旧テーブル構造に依存しているため、大幅な修正が必要。
- **トランザクション管理**: 数万件規模のインポートが想定されるため、SQLite の書き込みロック中に UI からの読み取りがブロッキングされないよう、適切なトランザクション分割や WAL モードの利用を検討する必要がある。
- **DB マイグレーション**: 既存の `dictionary_entries` テーブルがある場合、それを古い形式としてどう扱うか。初期段階なのでテーブル削除＆再作成でも許容される想定。
