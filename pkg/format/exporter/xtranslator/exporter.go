package xtranslator

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	formatexporter "github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

type exporter struct {
	logger *slog.Logger
}

// NewExporter creates a new instance of xTranslator XML exporter.
func NewExporter() formatexporter.Exporter {
	return &exporter{
		logger: slog.Default().With("slice", "xtranslator_exporter"),
	}
}

// GenerateXML builds and writes xTranslator-compatible XML from workflow input DTO.
func (e *exporter) GenerateXML(ctx context.Context, input formatexporter.ExportInput) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionExport)()

	outputPath := strings.TrimSpace(input.OutputFilePath)
	if outputPath == "" {
		return fmt.Errorf("output_file_path is required")
	}

	e.logger.DebugContext(ctx, "starting XML export",
		slog.String("xml_output", outputPath),
		slog.Int("term_count", len(input.TermResults)),
		slog.Int("main_count", len(input.MainResults)),
	)

	records, duplicateCount := mergeRecords(input.TermResults, input.MainResults)
	if duplicateCount > 0 {
		e.logger.WarnContext(ctx, "duplicate export records detected; later records override earlier ones",
			slog.Int("duplicate_count", duplicateCount),
		)
	}

	stringsList := make([]String, 0, len(records))
	for _, record := range records {
		stringsList = append(stringsList, String{
			SID:    e.formatSID(record.FormID),
			EDID:   strings.TrimSpace(record.EditorID),
			REC:    e.formatREC(record.RecordType),
			Source: record.SourceText,
			Dest:   record.TranslatedText,
		})
	}

	xmlRoot := SSTXMLRessources{
		Params: Params{
			Addon: strings.TrimSpace(input.PluginName),
		},
		Strings: stringsList,
	}

	output, err := xml.MarshalIndent(xmlRoot, "", "  ")
	if err != nil {
		e.logger.ErrorContext(ctx, "failed to marshal XML", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("marshal xtranslator XML: %w", err)
	}

	finalOutput := append([]byte(xml.Header), output...)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		e.logger.ErrorContext(ctx, "failed to create export directory", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("create output directory path=%s: %w", outputPath, err)
	}

	if err := os.WriteFile(outputPath, finalOutput, 0600); err != nil {
		e.logger.ErrorContext(ctx, "failed to write XML file", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("write XML output path=%s: %w", outputPath, err)
	}

	e.logger.InfoContext(ctx, "XML export completed",
		slog.String("path", outputPath),
		slog.Int("record_count", len(stringsList)),
		slog.String("plugin", xmlRoot.Params.Addon),
	)
	return nil
}

func mergeRecords(termResults, mainResults []formatexporter.ExportRecord) ([]formatexporter.ExportRecord, int) {
	merged := make([]formatexporter.ExportRecord, 0, len(termResults)+len(mainResults))
	indexByKey := make(map[string]int, len(termResults)+len(mainResults))
	duplicateCount := 0

	appendWithOverride := func(records []formatexporter.ExportRecord) {
		for _, record := range records {
			key := buildRecordKey(record)
			if idx, exists := indexByKey[key]; exists {
				merged[idx] = record
				duplicateCount++
				continue
			}
			indexByKey[key] = len(merged)
			merged = append(merged, record)
		}
	}

	appendWithOverride(termResults)
	appendWithOverride(mainResults)
	return merged, duplicateCount
}

func buildRecordKey(record formatexporter.ExportRecord) string {
	edid := strings.ToUpper(strings.TrimSpace(record.EditorID))
	rec := strings.ToUpper(strings.TrimSpace(record.RecordType))
	if edid == "" && rec == "" {
		return strings.ToUpper(strings.TrimSpace(record.FormID))
	}
	return edid + "|" + rec
}

// formatSID extracts 8-digit hex ID from "0x001234|Skyrim.esm".
func (e *exporter) formatSID(fullID string) string {
	hexPart := strings.TrimSpace(fullID)
	if hexPart == "" {
		return ""
	}

	parts := strings.Split(hexPart, "|")
	hexPart = strings.TrimPrefix(parts[0], "0x")
	hexPart = strings.TrimPrefix(hexPart, "0X")

	if len(hexPart) < 8 {
		hexPart = strings.Repeat("0", 8-len(hexPart)) + hexPart
	} else if len(hexPart) > 8 {
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
	return strings.TrimSpace(recordType)
}
