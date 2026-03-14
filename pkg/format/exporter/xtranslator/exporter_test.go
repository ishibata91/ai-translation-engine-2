package xtranslator

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	formatexporter "github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter"
)

func TestXTranslatorExporter_formatSID(t *testing.T) {
	e := &exporter{}
	tests := []struct {
		input    string
		expected string
	}{
		{"0x001234|Skyrim.esm", "00001234"},
		{"0xABCDEF|Dawnguard.esm", "00ABCDEF"},
		{"0x123|Update.esm", "00000123"},
		{"00001234", "00001234"},
		{"", ""},
	}

	for _, tt := range tests {
		result := e.formatSID(tt.input)
		if result != tt.expected {
			t.Errorf("formatSID(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestXTranslatorExporter_formatREC(t *testing.T) {
	e := &exporter{}
	tests := []struct {
		input    string
		expected string
	}{
		{"INFO NAM1", "INFO:NAM1"},
		{"QUST CNAM", "QUST:CNAM"},
		{"DIAL FULL", "DIAL:FULL"},
		{"NPC_", "NPC_"},
	}

	for _, tt := range tests {
		result := e.formatREC(tt.input)
		if result != tt.expected {
			t.Errorf("formatREC(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestXTranslatorExporter_GenerateXML(t *testing.T) {
	tmpDir := t.TempDir()
	xmlOutput := filepath.Join(tmpDir, "xtranslator.xml")

	e := NewExporter()
	err := e.GenerateXML(context.Background(), formatexporter.ExportInput{
		PluginName:     "Skyrim.esm",
		OutputFilePath: xmlOutput,
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
		t.Fatalf("GenerateXML failed: %v", err)
	}

	xmlData, err := os.ReadFile(xmlOutput)
	if err != nil {
		t.Fatal(err)
	}

	var root SSTXMLRessources
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		t.Fatalf("failed to unmarshal generated XML: %v", err)
	}

	if root.Params.Addon != "Skyrim.esm" {
		t.Errorf("expected Addon Skyrim.esm, got %s", root.Params.Addon)
	}

	if len(root.Strings) != 1 {
		t.Fatalf("expected 1 String element, got %d", len(root.Strings))
	}

	s := root.Strings[0]
	if s.SID != "00001234" {
		t.Errorf("expected sID 00001234, got %s", s.SID)
	}
	if s.REC != "INFO:NAM1" {
		t.Errorf("expected REC INFO:NAM1, got %s", s.REC)
	}
	if s.Source != "Hello world" {
		t.Errorf("expected Source Hello world, got %s", s.Source)
	}
	if s.Dest != "こんにちは世界" {
		t.Errorf("expected Dest こんにちは世界, got %s", s.Dest)
	}
}

func TestXTranslatorExporter_GenerateXML_DuplicateRecordPrefersMainResult(t *testing.T) {
	tmpDir := t.TempDir()
	xmlOutput := filepath.Join(tmpDir, "xtranslator-duplicate.xml")

	e := NewExporter()
	err := e.GenerateXML(context.Background(), formatexporter.ExportInput{
		PluginName:     "Skyrim.esm",
		OutputFilePath: xmlOutput,
		TermResults: []formatexporter.ExportRecord{
			{
				FormID:         "0x001234|Skyrim.esm",
				EditorID:       "DialogueGreeting",
				RecordType:     "INFO NAM1",
				SourceText:     "Hello world",
				TranslatedText: "こんにちは世界(用語)",
			},
		},
		MainResults: []formatexporter.ExportRecord{
			{
				FormID:         "0x001234|Skyrim.esm",
				EditorID:       "DialogueGreeting",
				RecordType:     "INFO NAM1",
				SourceText:     "Hello world",
				TranslatedText: "こんにちは世界(本文)",
			},
		},
	})
	if err != nil {
		t.Fatalf("GenerateXML failed: %v", err)
	}

	xmlData, err := os.ReadFile(xmlOutput)
	if err != nil {
		t.Fatal(err)
	}

	var root SSTXMLRessources
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		t.Fatalf("failed to unmarshal generated XML: %v", err)
	}

	if len(root.Strings) != 1 {
		t.Fatalf("expected 1 String element, got %d", len(root.Strings))
	}

	if root.Strings[0].Dest != "こんにちは世界(本文)" {
		t.Errorf("expected main result to override duplicate, got %s", root.Strings[0].Dest)
	}
}
