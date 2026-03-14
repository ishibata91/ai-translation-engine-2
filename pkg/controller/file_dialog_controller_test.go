package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestFileDialogController_API_TableDriven(t *testing.T) {
	testCases := []struct {
		name string
		run  func(t *testing.T, controller *FileDialogController)
	}{
		{
			name: "SelectFiles uses expected filter and returns files",
			run: func(t *testing.T, controller *FileDialogController) {
				controller.openMultipleFilesDialog = func(_ context.Context, options runtime.OpenDialogOptions) ([]string, error) {
					assert.Equal(t, "インポートする辞書ファイルを選択", options.Title)
					require.Len(t, options.Filters, 2)
					assert.Equal(t, "*.xml", options.Filters[0].Pattern)
					return []string{"a.xml", "b.xml"}, nil
				}
				files, err := controller.SelectFiles()
				require.NoError(t, err)
				assert.Equal(t, []string{"a.xml", "b.xml"}, files)
			},
		},
		{
			name: "SelectFiles wraps runtime error",
			run: func(t *testing.T, controller *FileDialogController) {
				controller.openMultipleFilesDialog = func(_ context.Context, _ runtime.OpenDialogOptions) ([]string, error) {
					return nil, errors.New("runtime failed")
				}
				_, err := controller.SelectFiles()
				require.Error(t, err)
				assert.Contains(t, err.Error(), "open multiple files dialog")
				assert.Contains(t, err.Error(), "runtime failed")
			},
		},
		{
			name: "SelectJSONFile uses expected filter and returns file",
			run: func(t *testing.T, controller *FileDialogController) {
				controller.openFileDialog = func(_ context.Context, options runtime.OpenDialogOptions) (string, error) {
					assert.Equal(t, "JSONファイルを選択", options.Title)
					require.Len(t, options.Filters, 2)
					assert.Equal(t, "*.json", options.Filters[0].Pattern)
					return "input.json", nil
				}
				path, err := controller.SelectJSONFile()
				require.NoError(t, err)
				assert.Equal(t, "input.json", path)
			},
		},
		{
			name: "SelectJSONFile wraps runtime error",
			run: func(t *testing.T, controller *FileDialogController) {
				controller.openFileDialog = func(_ context.Context, _ runtime.OpenDialogOptions) (string, error) {
					return "", errors.New("dialog failed")
				}
				_, err := controller.SelectJSONFile()
				require.Error(t, err)
				assert.Contains(t, err.Error(), "open json file dialog")
				assert.Contains(t, err.Error(), "dialog failed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewFileDialogController()
			tc.run(t, controller)
		})
	}
}
