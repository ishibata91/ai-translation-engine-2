package dictionary

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

type xmlImporter struct {
	config   Config
	store    DictionaryStore
	notifier progress.ProgressNotifier
	logger   *slog.Logger
}

// NewImporter は DictionaryImporter の新しいインスタンスを生成する。
func NewImporter(config Config, store DictionaryStore, notifier progress.ProgressNotifier, logger *slog.Logger) DictionaryImporter {
	return &xmlImporter{
		config:   config,
		store:    store,
		notifier: notifier,
		logger:   logger.With("component", "DictionaryImporter"),
	}
}

// ImportXML は xTranslator XML を io.Reader からストリーミングパースし、
// 許可された名詞レコードを sourceID に紐付けて保存する。
// 処理中は progress パッケージを通じて進捗をフロントエンドに通知する。
func (i *xmlImporter) ImportXML(ctx context.Context, sourceID int64, fileName string, file io.Reader) (int, error) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.ImportXML", "sourceID", sourceID, "fileName", fileName)
	defer i.logger.DebugContext(ctx, "EXIT DictionaryImporter.ImportXML")

	// dlc_sources を IMPORTING 状態に更新
	if err := i.store.UpdateSourceStatus(ctx, sourceID, "IMPORTING", 0, ""); err != nil {
		return 0, fmt.Errorf("failed to set source to IMPORTING: %w", err)
	}

	correlationID := fmt.Sprintf("dict-import-%d", sourceID)

	// 初回進捗通知
	i.notifier.OnProgress(ctx, progress.ProgressEvent{
		CorrelationID: correlationID,
		Status:        progress.StatusInProgress,
		Message:       fmt.Sprintf("辞書インポート開始: %s", fileName),
	})

	totalImported, err := i.parseAndSave(ctx, sourceID, correlationID, file)
	if err != nil {
		// エラー状態に更新して通知
		_ = i.store.UpdateSourceStatus(ctx, sourceID, "ERROR", totalImported, err.Error())
		i.notifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: correlationID,
			Completed:     totalImported,
			Status:        progress.StatusFailed,
			Message:       fmt.Sprintf("インポートエラー: %v", err),
		})
		return totalImported, err
	}

	// COMPLETED に更新
	if err := i.store.UpdateSourceStatus(ctx, sourceID, "COMPLETED", totalImported, ""); err != nil {
		return totalImported, fmt.Errorf("failed to set source to COMPLETED: %w", err)
	}

	// 完了通知
	i.notifier.OnProgress(ctx, progress.ProgressEvent{
		CorrelationID: correlationID,
		Completed:     totalImported,
		Status:        progress.StatusCompleted,
		Message:       fmt.Sprintf("インポート完了: %d 件", totalImported),
	})

	i.logger.InfoContext(ctx, "Successfully imported terms", "total", totalImported, "sourceID", sourceID)
	return totalImported, nil
}

// parseAndSave は XML を読み込み、バッチ単位で保存して合計件数を返す。
func (i *xmlImporter) parseAndSave(ctx context.Context, sourceID int64, correlationID string, file io.Reader) (int, error) {
	decoder := xml.NewDecoder(file)

	var addonName string
	const batchSize = 1000
	batch := make([]DictTerm, 0, batchSize)
	totalImported := 0

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return totalImported, fmt.Errorf("error reading xml token: %w", err)
		}
		if t == nil {
			break
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		switch se.Name.Local {
		case "Addon":
			addonName = i.handleAddonElement(ctx, decoder, &se, addonName)
		case "String":
			term, ok := i.handleStringElement(ctx, decoder, &se, sourceID)
			if !ok {
				continue
			}
			batch = append(batch, term)

			if len(batch) >= batchSize {
				flushed, err := i.flushBatch(ctx, batch)
				if err != nil {
					return totalImported, err
				}
				totalImported += flushed
				batch = batch[:0]

				// バッチ完了ごとに進捗通知
				i.notifier.OnProgress(ctx, progress.ProgressEvent{
					CorrelationID: correlationID,
					Completed:     totalImported,
					Status:        progress.StatusInProgress,
					Message:       fmt.Sprintf("インポート中: %d 件処理済み", totalImported),
				})
			}
		}
	}

	// 残りのバッチをフラッシュ
	if len(batch) > 0 {
		flushed, err := i.flushBatch(ctx, batch)
		if err != nil {
			return totalImported, err
		}
		totalImported += flushed
	}

	return totalImported, nil
}

// handleAddonElement は Addon 要素をデコードしてアドオン名を返す。
func (i *xmlImporter) handleAddonElement(ctx context.Context, decoder *xml.Decoder, se *xml.StartElement, currentAddon string) string {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.handleAddonElement")

	var addon string
	err := decoder.DecodeElement(&addon, se)
	if err == nil && currentAddon == "" {
		return addon
	}
	return currentAddon
}

// handleStringElement は String 要素をデコードして DictTerm を返す。
// REC タイプが許可されていない場合は (DictTerm{}, false) を返す。
func (i *xmlImporter) handleStringElement(ctx context.Context, decoder *xml.Decoder, se *xml.StartElement, sourceID int64) (DictTerm, bool) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.handleStringElement")

	var strElem struct {
		EDID   string `xml:"EDID"`
		REC    string `xml:"REC"`
		Source string `xml:"Source"`
		Dest   string `xml:"Dest"`
	}

	err := decoder.DecodeElement(&strElem, se)
	if err != nil {
		i.logger.WarnContext(ctx, "failed to decode String element, skipping", "error", err)
		return DictTerm{}, false
	}

	if !i.config.IsAllowedREC(strElem.REC) {
		return DictTerm{}, false
	}

	return DictTerm{
		SourceID:   sourceID,
		EDID:       strElem.EDID,
		RecordType: strElem.REC,
		Source:     strElem.Source,
		Dest:       strElem.Dest,
	}, true
}

// flushBatch はバッチのエントリをストアに保存し、保存件数を返す。
func (i *xmlImporter) flushBatch(ctx context.Context, batch []DictTerm) (int, error) {
	i.logger.DebugContext(ctx, "ENTER DictionaryImporter.flushBatch", slog.Int("batchSize", len(batch)))

	if err := i.store.SaveTerms(ctx, batch); err != nil {
		return 0, fmt.Errorf("error saving batch: %w", err)
	}
	return len(batch), nil
}
