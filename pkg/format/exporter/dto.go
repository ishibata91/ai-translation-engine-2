package exporter

// ExportInput is the slice-local DTO used by workflow to request xTranslator XML generation.
type ExportInput struct {
	PluginName     string
	SourceLanguage string
	DestLanguage   string
	TermResults    []ExportRecord
	MainResults    []ExportRecord
	OutputFilePath string
}

// ExportRecord is one normalized export row used to build XML nodes.
type ExportRecord struct {
	FormID         string
	EditorID       string
	RecordType     string
	SourceText     string
	TranslatedText string
}
