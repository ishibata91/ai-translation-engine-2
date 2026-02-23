package export

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"os"
	"testing"
)

func TestExporter_formatSID(t *testing.T) {
	e := &exporter{}
	tests := []struct {
		input    string
		expected string
	}{
		{"0x001234|Skyrim.esm", "00001234"},
		{"0xABCDEF|Dawnguard.esm", "00ABCDEF"},
		{"0x123|Update.esm", "00000123"},
		{"00001234", "00001234"},
	}

	for _, tt := range tests {
		result := e.formatSID(tt.input)
		if result != tt.expected {
			t.Errorf("formatSID(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestExporter_formatREC(t *testing.T) {
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

func TestExporter_ExportToXML(t *testing.T) {
	// Create temporary JSON file
	jsonContent := `[
		{
			"id": "0x001234|Skyrim.esm",
			"type": "INFO NAM1",
			"source_text": "Hello world",
			"translated_text": "こんにちは世界",
			"source_plugin": "Skyrim.esm",
			"editor_id": "DialogueGreeting"
		}
	]`
	tmpJSON, err := ioutil.TempFile("", "test_export_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpJSON.Name())
	tmpJSON.WriteString(jsonContent)
	tmpJSON.Close()

	xmlOutput := tmpJSON.Name() + ".xml"
	defer os.Remove(xmlOutput)

	e := NewExporter()
	err = e.ExportToXML(context.Background(), tmpJSON.Name(), xmlOutput)
	if err != nil {
		t.Fatalf("ExportToXML failed: %v", err)
	}

	// Read and verify XML
	xmlData, err := ioutil.ReadFile(xmlOutput)
	if err != nil {
		t.Fatal(err)
	}

	var root SSTXMLRessources
	err = xml.Unmarshal(xmlData, &root)
	if err != nil {
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
