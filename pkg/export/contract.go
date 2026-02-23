package export

import "context"

// Exporter is the interface for ExportSlice, responsible for generating xTranslator XML.
type Exporter interface {
	ExportToXML(ctx context.Context, jsonPath string, xmlOutputPath string) error
}
