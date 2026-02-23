package export_slice

// InputJSON represents the structure of the JSON file output by Pass 2.
type InputJSON []TranslationResult

// TranslationResult matches the structure expected from Pass 2.
type TranslationResult struct {
	ID             string  `json:"id"`
	RecordType     string  `json:"type"`
	SourceText     string  `json:"source_text"`
	TranslatedText *string `json:"translated_text,omitempty"`
	Index          *int    `json:"index,omitempty"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	SourcePlugin   string  `json:"source_plugin"`
	SourceFile     string  `json:"source_file"`
	EditorID       *string `json:"editor_id,omitempty"`
	ParentID       *string `json:"parent_id,omitempty"`
	ParentEditorID *string `json:"parent_editor_id,omitempty"`
}

// SSTXMLRessources is the root element for xTranslator XML.
type SSTXMLRessources struct {
	Params  Params   `xml:"Params"`
	Strings []String `xml:"Strings>String"`
}

type Params struct {
	Addon string `xml:"Addon,attr"`
}

type String struct {
	SID    string `xml:"sID,attr"`
	EDID   string `xml:"EDID"`
	REC    string `xml:"REC"`
	Source string `xml:"Source"`
	Dest   string `xml:"Dest"`
}
