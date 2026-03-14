package export_test

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	formatexporter "github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter/xtranslator"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
)

func TestXMLExportService_GenerateXTranslatorXML(t *testing.T) {
	service := workflow.NewXMLExportService(xtranslator.NewExporter())

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "workflow-export.xml")

	err := service.GenerateXTranslatorXML(context.Background(), formatexporter.ExportInput{
		PluginName:     "Skyrim.esm",
		OutputFilePath: outputPath,
		MainResults: []formatexporter.ExportRecord{
			{
				FormID:         "0x001234|Skyrim.esm",
				EditorID:       "DialogueGreeting",
				RecordType:     "INFO NAM1",
				SourceText:     "Hello world",
				TranslatedText: "こんにちは世界",
			},
		},
	})
	if err != nil {
		t.Fatalf("GenerateXTranslatorXML failed: %v", err)
	}

	xmlBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read XML output failed: %v", err)
	}

	var root xtranslator.SSTXMLRessources
	if err := xml.Unmarshal(xmlBytes, &root); err != nil {
		t.Fatalf("unmarshal XML output failed: %v", err)
	}

	if root.Params.Addon != "Skyrim.esm" {
		t.Fatalf("expected plugin Skyrim.esm, got %s", root.Params.Addon)
	}
	if len(root.Strings) != 1 {
		t.Fatalf("expected 1 String element, got %d", len(root.Strings))
	}
	if root.Strings[0].REC != "INFO:NAM1" {
		t.Fatalf("expected REC INFO:NAM1, got %s", root.Strings[0].REC)
	}
}
