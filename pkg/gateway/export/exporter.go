package export

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

type exporter struct {
	logger *slog.Logger
}

// NewExporter creates a new instance of xTranslator XML exporter.
func NewExporter() Exporter {
	return &exporter{
		logger: slog.Default().With("slice", "ExportSlice"),
	}
}

func (e *exporter) ExportToXML(ctx context.Context, jsonPath string, xmlOutputPath string) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionExport)()
	e.logger.DebugContext(ctx, "starting XML export", slog.String("json_path", jsonPath), slog.String("xml_output", xmlOutputPath))

	// 1. Read JSON
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		e.logger.ErrorContext(ctx, "failed to read JSON file for export", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var results []TranslationResult
	if err := json.Unmarshal(content, &results); err != nil {
		e.logger.ErrorContext(ctx, "failed to parse JSON for export", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 2. Map to XML structure
	var stringsList []String
	pluginName := ""

	for _, res := range results {
		if pluginName == "" && res.SourcePlugin != "" {
			pluginName = res.SourcePlugin
		}

		// Extract sID (8-digit hex)
		sID := e.formatSID(res.ID)

		// Map Record Type (signature)
		rec := e.formatREC(res.RecordType)

		// Determine EDID
		edid := ""
		if res.EditorID != nil {
			edid = *res.EditorID
		}
		if res.ParentEditorID != nil && *res.ParentEditorID != "" {
			edid = *res.ParentEditorID
		}

		dest := ""
		if res.TranslatedText != nil {
			dest = *res.TranslatedText
		}

		stringsList = append(stringsList, String{
			SID:    sID,
			EDID:   edid,
			REC:    rec,
			Source: res.SourceText,
			Dest:   dest,
		})
	}

	xmlRoot := SSTXMLRessources{
		Params: Params{
			Addon: pluginName,
		},
		Strings: stringsList,
	}

	// 3. Generate XML
	output, err := xml.MarshalIndent(xmlRoot, "", "  ")
	if err != nil {
		e.logger.ErrorContext(ctx, "failed to marshal XML", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	header := []byte(xml.Header)
	finalOutput := append(header, output...)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(xmlOutputPath), 0755); err != nil {
		e.logger.ErrorContext(ctx, "failed to create export directory", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(xmlOutputPath, finalOutput, 0600); err != nil {
		e.logger.ErrorContext(ctx, "failed to write XML file", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("failed to write XML file: %w", err)
	}

	e.logger.InfoContext(ctx, "XML export completed",
		slog.String("path", xmlOutputPath),
		slog.Int("record_count", len(stringsList)),
		slog.String("plugin", pluginName),
	)
	return nil
}

// formatSID extracts 8-digit hex ID from "0x001234|Skyrim.esm".
func (e *exporter) formatSID(fullID string) string {
	// Split by |
	parts := strings.Split(fullID, "|")
	hexPart := parts[0]

	// Remove 0x prefix
	hexPart = strings.TrimPrefix(hexPart, "0x")

	// Pad to 8 digits
	if len(hexPart) < 8 {
		hexPart = strings.Repeat("0", 8-len(hexPart)) + hexPart
	} else if len(hexPart) > 8 {
		// If it's something like 00001234 (already 8 digits but maybe with leading zeros or esl/esp bits)
		hexPart = hexPart[len(hexPart)-8:]
	}

	return strings.ToUpper(hexPart)
}

// formatREC converts "INFO NAM1" to "INFO:NAM1".
func (e *exporter) formatREC(recordType string) string {
	parts := strings.Fields(recordType)
	if len(parts) >= 2 {
		return fmt.Sprintf("%s:%s", parts[0], parts[1])
	}
	return recordType
}
