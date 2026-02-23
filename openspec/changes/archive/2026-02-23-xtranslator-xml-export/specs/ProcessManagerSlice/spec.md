## MODIFIED Requirements

### Requirement: Mapping to TranslationRequest
The Process Manager Slice SHALL map the generic DTO models produced by the Loader Slice into domain-specific `TranslationRequest` models optimized for the translation process, preserving parent identifiers for hierarchical records.

#### Scenario: Map hierarchical records to TranslationRequests
- **WHEN** unfolding hierarchical records such as Quest Stages or Quest Objectives
- **THEN** ProcessManagerSlice SHALL inherit and copy the parent record's `ID` (FormID) and `EditorID` to the child's `TranslationRequest`
- **AND** this behavior SHALL ensure that key metadata is not lost for downstream XML generation by the Export Slice
