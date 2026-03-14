package exporter

import "context"

// Exporter defines the workflow-facing contract for XML export generation.
type Exporter interface {
	GenerateXML(ctx context.Context, input ExportInput) error
}
