package workflow

import (
	"context"
	"fmt"

	formatexporter "github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter"
)

// XMLExportService is a workflow adapter that delegates XML generation to format/exporter contract.
type XMLExportService struct {
	exporter formatexporter.Exporter
}

// NewXMLExportService constructs a workflow service bound to Exporter contract.
func NewXMLExportService(exporter formatexporter.Exporter) *XMLExportService {
	return &XMLExportService{exporter: exporter}
}

// GenerateXTranslatorXML exports workflow output into xTranslator XML through the Exporter contract.
func (s *XMLExportService) GenerateXTranslatorXML(ctx context.Context, input formatexporter.ExportInput) error {
	if s.exporter == nil {
		return fmt.Errorf("xml exporter is not configured")
	}
	return s.exporter.GenerateXML(ctx, input)
}
