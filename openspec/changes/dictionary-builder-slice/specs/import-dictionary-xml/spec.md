# Import Dictionary XML Specification

## ADDED Requirements

### Requirement: Import xTranslator XML
The system MUST be able to read one or more xTranslator XML files and extract noun records based on a configurable allowlist.

#### Scenario: Valid XML with mixed records is imported
- **WHEN** the user provides a valid xTranslator XML file containing `BOOK:FULL`, `NPC_:FULL`, and `INFO` records
- **AND** the configuration specifies `BOOK:FULL` and `NPC_:FULL` as allowed types
- **THEN** the system extracts only the `BOOK:FULL` and `NPC_:FULL` records
- **AND** ignores the `INFO` records
- **AND** upserts the extracted records into the SQLite dictionary database

#### Scenario: XML file is extremely large
- **WHEN** the user provides an XML file that exceeds available memory
- **THEN** the system must parse the file using a streaming approach (`encoding/xml.Decoder`) without causing an Out Of Memory (OOM) error

#### Scenario: Database schema does not exist
- **WHEN** the application starts up and initializes the `DictionaryStore`
- **THEN** the system automatically executes the `CREATE TABLE IF NOT EXISTS` statement to ensure the dictionary tables exist before any imports over

#### Scenario: Duplicate records are imported
- **WHEN** the user imports an XML file containing a record with an EDID that already exists in the database
- **THEN** the system executes an UPSERT (`ON CONFLICT`) operation
- **AND** updates the existing record with the new `Source` and `Dest` values without creating duplicate rows
