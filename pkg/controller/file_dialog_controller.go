package controller

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type openFileDialogFunc func(ctx context.Context, options runtime.OpenDialogOptions) (string, error)
type openMultipleFilesDialogFunc func(ctx context.Context, options runtime.OpenDialogOptions) ([]string, error)

// FileDialogController exposes Wails-facing file selection dialogs.
type FileDialogController struct {
	ctx                     context.Context
	openFileDialog          openFileDialogFunc
	openMultipleFilesDialog openMultipleFilesDialogFunc
}

// NewFileDialogController constructs the file dialog controller adapter.
func NewFileDialogController() *FileDialogController {
	return &FileDialogController{
		ctx:                     context.Background(),
		openFileDialog:          runtime.OpenFileDialog,
		openMultipleFilesDialog: runtime.OpenMultipleFilesDialog,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *FileDialogController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// SelectFiles opens a multi-file dialog for dictionary import files.
func (c *FileDialogController) SelectFiles() ([]string, error) {
	files, err := c.openMultipleFilesDialog(c.context(), runtime.OpenDialogOptions{
		Title: "インポートする辞書ファイルを選択",
		Filters: []runtime.FileFilter{
			{DisplayName: "XML Files (*.xml)", Pattern: "*.xml"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open multiple files dialog: %w", err)
	}
	return files, nil
}

// SelectTranslationInputFiles opens a multi-file dialog for translation input JSON files.
func (c *FileDialogController) SelectTranslationInputFiles() ([]string, error) {
	files, err := c.openMultipleFilesDialog(c.context(), runtime.OpenDialogOptions{
		Title: "翻訳対象JSONファイルを選択",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open translation input files dialog: %w", err)
	}
	return files, nil
}

// SelectJSONFile opens a single-file dialog for JSON input.
func (c *FileDialogController) SelectJSONFile() (string, error) {
	path, err := c.openFileDialog(c.context(), runtime.OpenDialogOptions{
		Title: "JSONファイルを選択",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("open json file dialog: %w", err)
	}
	return path, nil
}

func (c *FileDialogController) context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}
