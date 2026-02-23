## ADDED Requirements

### Requirement: XML File Generation
ExportSlice SHALL generate an XML file compatible with xTranslator (SSTXMLRessources) from the Pass 2 output JSON.

#### Scenario: Generate XML from translated JSON
- **WHEN** the system receives the JSON output from Pass 2
- **THEN** ExportSlice SHALL parse the JSON
- **AND** generate an xTranslator-compatible XML file

### Requirement: sID Formatting
ExportSlice SHALL extract a plain 8-digit hex ID from the load-order independent `sID` format.

#### Scenario: Extract sID
- **WHEN** processing an entry with `sID` like `0x001234|Skyrim.esm`
- **THEN** ExportSlice SHALL extract `00001234`
- **AND** apply it to the XML attribute `<String sID="...">`

### Requirement: Tag Mapping
ExportSlice SHALL correctly map JSON fields to corresponding XML tags based on predefined rules.

#### Scenario: Map fields to tags
- **WHEN** generating the XML output
- **THEN** the `EditorID` field SHALL be mapped to `<EDID>`
- **AND** the `RecordType` field SHALL be mapped to `<REC>` (e.g., `INFO NAM1` -> `INFO:NAM1`)
- **AND** the `SourceText` field SHALL be mapped to `<Source>`
- **AND** the `TranslatedText` field SHALL be mapped to `<Dest>`
