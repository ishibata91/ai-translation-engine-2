## 1. Setup and Pre-conditions

- [ ] 1.1 Update `extractData.pas` to separate `stage_index` and `log_index`, and output explicit `order` for conversation responses.

## 2. DTO and Metadata Propagation

- [ ] 2.1 Update `LoaderSlice` DTOs (e.g., `QuestStage`, `QuestObjective`) to include fields for parent `ID` (FormID) and `EditorID`.
- [ ] 2.2 Update `LoaderSlice` parsing logic to correctly parse separated indices and map parent metadata to child nodes.
- [ ] 2.3 Update `ProcessManagerSlice` mapping logic to inherit and copy parent `ID` and `EditorID` into the `TranslationRequest` structure when unfolding hierarchical records.

## 3. Translation Output Formatting

- [ ] 3.1 Update `Pass2TranslatorSlice` JSON serialization to preserve the full signature (e.g., `INFO NAM1`) in the `type` field, removing any logic that truncates it.

## 4. Export Slice Implementation

- [ ] 4.1 Create the new `ExportSlice` component structure.
- [ ] 4.2 Implement JSON parsing logic in `ExportSlice` to read the translated output from Pass 2.
- [ ] 4.3 Implement `sID` formatting logic (extracting 8-digit hex ID from `0x001234|Skyrim.esm`).
- [ ] 4.4 Implement tag mapping logic (EditorID -> `<EDID>`, RecordType -> `<REC>`, SourceText -> `<Source>`, TranslatedText -> `<Dest>`).
- [ ] 4.5 Implement XML generation logic adhering to the `SSTXMLRessources` xTranslator format.

## 5. Verification

- [ ] 5.1 Write unit tests for `ExportSlice` parsing and XML generation.
- [ ] 5.2 Validate that the generated XML can be successfully imported by xTranslator without errors.
- [ ] 5.3 Run end-to-end integration test validating the entire data propagation flow from Pascal extraction to XML output.
