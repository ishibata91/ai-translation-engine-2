package dictionary

// DictTerm represents a single dictionary entry extracted from xTranslator XML.
type DictTerm struct {
	EDID   string `json:"edid"`
	REC    string `json:"rec"`
	Source string `json:"source"`
	Dest   string `json:"dest"`
	Addon  string `json:"addon"`
}
