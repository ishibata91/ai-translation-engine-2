## MODIFIED Requirements

### Requirement: JSON Output Formatting
Pass2TranslatorSlice SHALL format the translated JSON output containing original context alongside translations without truncating data types.

#### Scenario: Save full signature format
- **WHEN** generating the final JSON output
- **THEN** Pass2TranslatorSlice SHALL preserve the full signature in the `type` field (e.g., `INFO NAM1` or `QUST CNAM`)
- **AND** SHALL NOT round or truncate the `type` field to a generic record type like `INFO` or `QUST`
