package dictionary

import (
	"context"
	"io"
)

// DictionaryImporter は XML のパースと辞書の永続化をオーケストレートする。
type DictionaryImporter interface {
	// ImportXML は XML ファイルを読み込み、sourceID に紐付けてエントリを保存する。
	// ファイルのメタデータ（fileName, fileSize）を受け取り、dlc_sources のライフサイクルを管理する。
	ImportXML(ctx context.Context, sourceID int64, fileName string, file io.Reader) (int, error)
}

// DictionaryStore は SQLite への辞書データ永続化を担う。
// このスライスはテーブル作成・INSERT/UPSERT・CRUD 操作すべてを所有する。
type DictionaryStore interface {
	// --- 辞書ソース管理 ---

	// GetSources は dlc_sources の全レコードを返す。
	GetSources(ctx context.Context) ([]DictSource, error)

	// CreateSource は新しい辞書ソースレコードを作成し、採番された ID を返す。
	CreateSource(ctx context.Context, src *DictSource) (int64, error)

	// UpdateSourceStatus は指定ソースのステータス・エントリ数・エラーメッセージを更新する。
	UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error

	// DeleteSource は指定ソースを削除する（関連エントリはカスケード削除）。
	DeleteSource(ctx context.Context, id int64) error

	// --- 辞書エントリ管理 ---

	// GetEntriesBySourceID は指定ソースに紐付く全エントリを返す（後方互換用）。
	GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]DictTerm, error)

	// GetEntriesBySourceIDPaginated は指定ソースのエントリをページネーション付きで返す。
	GetEntriesBySourceIDPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, limit, offset int) (*DictTermPage, error)

	// SearchAllEntriesPaginated は全ソースを横断してエントリを検索する。
	// 各エントリには SourceName (dlc_sources.file_name) が付与される。
	SearchAllEntriesPaginated(ctx context.Context, query string, filters map[string]string, limit, offset int) (*DictTermPage, error)

	// SaveTerms は複数エントリをバッチで挿入する。
	SaveTerms(ctx context.Context, terms []DictTerm) error

	// UpdateEntry は指定エントリの source_text / dest_text を更新する。
	UpdateEntry(ctx context.Context, term DictTerm) error

	// DeleteEntry は指定エントリを削除する。
	DeleteEntry(ctx context.Context, id int64) error
}
