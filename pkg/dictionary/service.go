package dictionary

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// DictionaryService は UI レベルのアクションをオーケストレートする Wails 向けサービス。
// DictionaryStore と DictionaryImporter の呼び出しをまとめ、フロントエンドに単一の入り口を提供する。
type DictionaryService struct {
	store    DictionaryStore
	importer DictionaryImporter
	logger   *slog.Logger
}

// NewDictionaryService は DictionaryService の新しいインスタンスを生成する。
func NewDictionaryService(store DictionaryStore, importer DictionaryImporter, logger *slog.Logger) *DictionaryService {
	return &DictionaryService{
		store:    store,
		importer: importer,
		logger:   logger.With("component", "DictionaryService"),
	}
}

// GetSources は登録済みの辞書ソース一覧を返す。
func (s *DictionaryService) GetSources(ctx context.Context) ([]DictSource, error) {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.GetSources")
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.GetSources")
	return s.store.GetSources(ctx)
}

// DeleteSource は指定された辞書ソースとその全エントリを削除する。
func (s *DictionaryService) DeleteSource(ctx context.Context, id int64) error {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.DeleteSource", "id", id)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.DeleteSource")
	return s.store.DeleteSource(ctx, id)
}

// GetEntries は指定ソースに紐付く辞書エントリ一覧を返す（後方互換用）。
func (s *DictionaryService) GetEntries(ctx context.Context, sourceID int64) ([]DictTerm, error) {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.GetEntries", "sourceID", sourceID)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.GetEntries")
	return s.store.GetEntriesBySourceID(ctx, sourceID)
}

// GetEntriesPaginated は指定ソースのエントリをページネーション付きで返す。
// page は1始まり、pageSize は取得件数（例: 500）。
func (s *DictionaryService) GetEntriesPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, page, pageSize int) (*DictTermPage, error) {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.GetEntriesPaginated", "sourceID", sourceID, "query", query, "filters", filters, "page", page, "pageSize", pageSize)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.GetEntriesPaginated")
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.store.GetEntriesBySourceIDPaginated(ctx, sourceID, query, filters, pageSize, offset)
}

// SearchAll は全辞書ソースを横断してエントリを検索する。
// page は1始まり、pageSize は取得件数。
func (s *DictionaryService) SearchAll(ctx context.Context, query string, filters map[string]string, page, pageSize int) (*DictTermPage, error) {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.SearchAll", "query", query, "filters", filters, "page", page, "pageSize", pageSize)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.SearchAll")
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.store.SearchAllEntriesPaginated(ctx, query, filters, pageSize, offset)
}

// UpdateEntry は指定エントリの source_text / dest_text を更新する。
func (s *DictionaryService) UpdateEntry(ctx context.Context, term DictTerm) error {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.UpdateEntry", "id", term.ID)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.UpdateEntry")
	return s.store.UpdateEntry(ctx, term)
}

// DeleteEntry は指定エントリを削除する。
func (s *DictionaryService) DeleteEntry(ctx context.Context, id int64) error {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.DeleteEntry", "id", id)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.DeleteEntry")
	return s.store.DeleteEntry(ctx, id)
}

// StartImport は指定ファイルのインポートを開始する。
// dlc_sources に PENDING レコードを作成した後、非同期でインポート処理を実行する。
// 戻り値は作成されたソースの ID。
func (s *DictionaryService) StartImport(ctx context.Context, filePath string) (int64, error) {
	s.logger.DebugContext(ctx, "ENTER DictionaryService.StartImport", "filePath", filePath)
	defer s.logger.DebugContext(ctx, "EXIT DictionaryService.StartImport")

	// ファイル情報を取得
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	// dlc_sources に PENDING レコードを作成
	src := &DictSource{
		FileName: filepath.Base(filePath),
		Format:   "xml",
		FilePath: filePath,
		FileSize: stat.Size(),
		Status:   "PENDING",
	}
	sourceID, err := s.store.CreateSource(ctx, src)
	if err != nil {
		return 0, fmt.Errorf("failed to create source record: %w", err)
	}

	// 非同期でインポート実行
	go func() {
		bgCtx := context.Background()
		file, err := os.Open(filePath)
		if err != nil {
			s.logger.ErrorContext(bgCtx, "failed to open import file", "error", err, "sourceID", sourceID)
			_ = s.store.UpdateSourceStatus(bgCtx, sourceID, "ERROR", 0, err.Error())
			return
		}
		defer file.Close()

		count, err := s.importer.ImportXML(bgCtx, sourceID, src.FileName, file)
		if err != nil {
			s.logger.ErrorContext(bgCtx, "import failed", "error", err, "sourceID", sourceID, "count", count)
			// UpdateSourceStatus は importer.go 内で呼ばれるため、ここでは追加の処理不要
		}
	}()

	return sourceID, nil
}
