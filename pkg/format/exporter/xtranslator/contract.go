package xtranslator

import formatexporter "github.com/ishibata91/ai-translation-engine-2/pkg/format/exporter"

// Exporter is the workflow-facing XML export contract.
type Exporter = formatexporter.Exporter

// ExportInput is the workflow-facing input DTO for XML export.
type ExportInput = formatexporter.ExportInput

// ExportRecord is one normalized record for XML export.
type ExportRecord = formatexporter.ExportRecord
