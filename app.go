package main

import (
	"context"
	"fmt"

	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	dictService *dictionary.DictionaryService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// SetDictService sets the dictionary service instance
func (a *App) SetDictService(dictService *dictionary.DictionaryService) {
	a.dictService = dictService
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform cleanup operations here if needed
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// ── 辞書ビルダー用 API ラッパー ──────────────────────────

// DictGetSources は登録済みの辞書ソース一覧を返す。
func (a *App) DictGetSources() ([]dictionary.DictSource, error) {
	return a.dictService.GetSources(a.ctx)
}

// DictDeleteSource は指定された辞書ソースとその全エントリを削除する。
func (a *App) DictDeleteSource(id int64) error {
	return a.dictService.DeleteSource(a.ctx, id)
}

// DictGetEntries は指定ソースに紐付く辞書エントリ一覧を返す（後方互換用）。
func (a *App) DictGetEntries(sourceID int64) ([]dictionary.DictTerm, error) {
	return a.dictService.GetEntries(a.ctx, sourceID)
}

// DictGetEntriesPaginated は指定ソースのエントリをページネーション付きで返す。
// page は1始まり、pageSize は取得件数（例: 500）、query は検索キーワード（空文字で全件）。
func (a *App) DictGetEntriesPaginated(sourceID int64, query string, filters map[string]string, page, pageSize int) (*dictionary.DictTermPage, error) {
	return a.dictService.GetEntriesPaginated(a.ctx, sourceID, query, filters, page, pageSize)
}

// DictSearchAllEntriesPaginated は全辞書ソースを横断してエントリを検索する。
func (a *App) DictSearchAllEntriesPaginated(query string, filters map[string]string, page, pageSize int) (*dictionary.DictTermPage, error) {
	return a.dictService.SearchAll(a.ctx, query, filters, page, pageSize)
}

// DictUpdateEntry は指定エントリの source_text / dest_text を更新する。
func (a *App) DictUpdateEntry(term dictionary.DictTerm) error {
	return a.dictService.UpdateEntry(a.ctx, term)
}

// DictDeleteEntry は指定エントリを削除する。
func (a *App) DictDeleteEntry(id int64) error {
	return a.dictService.DeleteEntry(a.ctx, id)
}

// DictStartImport は指定ファイルのインポートを開始する。
func (a *App) DictStartImport(filePath string) (int64, error) {
	return a.dictService.StartImport(a.ctx, filePath)
}

// SelectFiles はOSのファイル選択ダイアログを開き、選択されたファイルの絶対パス一覧を返す。
func (a *App) SelectFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "インポートする辞書ファイルを選択",
		Filters: []runtime.FileFilter{
			{DisplayName: "XML Files (*.xml)", Pattern: "*.xml"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
}
