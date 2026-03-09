package dictionary

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/telemetry"
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
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "fetching dictionary sources")
	return s.store.GetSources(ctx)
}

// DeleteSource は指定された辞書ソースとその全エントリを削除する。
func (s *DictionaryService) DeleteSource(ctx context.Context, id int64) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.InfoContext(ctx, "deleting dictionary source", slog.Int64("id", id))
	return s.store.DeleteSource(ctx, id)
}

// GetEntries は指定ソースに紐付く辞書エントリ一覧を返す（後方互換用）。
func (s *DictionaryService) GetEntries(ctx context.Context, sourceID int64) ([]DictTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "fetching entries for source", slog.Int64("source_id", sourceID))
	return s.store.GetEntriesBySourceID(ctx, sourceID)
}

// GetEntriesPaginated は指定ソースのエントリをページネーション付きで返す。
// page は1始まり、pageSize は取得件数（例: 500）。
func (s *DictionaryService) GetEntriesPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, page, pageSize int) (*DictTermPage, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "fetching entries paginated",
		slog.Int64("source_id", sourceID),
		slog.String("query", query),
		slog.Int("page", page),
	)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.store.GetEntriesBySourceIDPaginated(ctx, sourceID, query, filters, pageSize, offset)
}

// SearchAll は全辞書ソースを横断してエントリを検索する。
// page は1始まり、pageSize は取得件数。
func (s *DictionaryService) SearchAll(ctx context.Context, query string, filters map[string]string, page, pageSize int) (*DictTermPage, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching all dictionaries",
		slog.String("query", query),
		slog.Int("page", page),
	)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.store.SearchAllEntriesPaginated(ctx, query, filters, pageSize, offset)
}

// UpdateEntry は指定エントリの source_text / dest_text を更新する。
func (s *DictionaryService) UpdateEntry(ctx context.Context, term DictTerm) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.InfoContext(ctx, "updating dictionary entry", slog.Int64("id", term.ID))
	return s.store.UpdateEntry(ctx, term)
}

// DeleteEntry は指定エントリを削除する。
func (s *DictionaryService) DeleteEntry(ctx context.Context, id int64) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.InfoContext(ctx, "deleting dictionary entry", slog.Int64("id", id))
	return s.store.DeleteEntry(ctx, id)
}

// StartImport は指定ファイルのインポートを開始する。
// dlc_sources に PENDING レコードを作成した後、非同期でインポート処理を実行する。
// 戻り値は作成されたソースの ID。
func (s *DictionaryService) StartImport(ctx context.Context, filePath string) (int64, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionImport)()
	s.logger.InfoContext(ctx, "starting dictionary import", slog.String("file_path", filePath))

	// ファイル情報を取得
	stat, err := os.Stat(filePath)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to stat file for import", telemetry2.ErrorAttrs(err)...)
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
		s.logger.ErrorContext(ctx, "failed to create source record", telemetry2.ErrorAttrs(err)...)
		return 0, fmt.Errorf("failed to create source record: %w", err)
	}

	// 非同期でインポート実行
	go func() {
		// リクエストIDを引き継ぐ
		bgCtx := telemetry2.WithAttrs(ctx, slog.String("request_id", "async-import-"+uuid.New().String()))
		defer telemetry2.StartSpan(bgCtx, telemetry2.ActionImport)()

		s.logger.InfoContext(bgCtx, "background import task started", slog.Int64("source_id", sourceID))

		file, err := os.Open(filePath)
		if err != nil {
			s.logger.ErrorContext(bgCtx, "failed to open import file",
				append(telemetry2.ErrorAttrs(err), slog.Int64("source_id", sourceID))...)
			_ = s.store.UpdateSourceStatus(bgCtx, sourceID, "ERROR", 0, err.Error())
			return
		}
		defer file.Close()

		count, err := s.importer.ImportXML(bgCtx, sourceID, src.FileName, file)
		if err != nil {
			s.logger.ErrorContext(bgCtx, "import process failed",
				append(telemetry2.ErrorAttrs(err), slog.Int64("source_id", sourceID), slog.Int("processed_count", count))...)
		} else {
			s.logger.InfoContext(bgCtx, "import process completed",
				slog.Int64("source_id", sourceID), slog.Int("processed_count", count))
		}
	}()

	return sourceID, nil
}
