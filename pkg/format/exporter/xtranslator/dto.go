package xtranslator

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
