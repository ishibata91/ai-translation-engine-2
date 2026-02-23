## MODIFIED Requirements

### Requirement: Loader DTO Expansion
The Loader Slice SHALL define and populate robust Data Transfer Objects (DTOs) that capture hierarchical data relationships necessary for full translation context and subsequent export processing.

#### Scenario: Populate DTO for hierarchical Quest items
- **WHEN** loading data from extraction scripts into DTOs
- **THEN** LoaderSlice SHALL expand the structures for `QuestStage` and `QuestObjective` to store their parent Quest's `ID` and `EditorID`
- **AND** populate these fields during the data loading process

### Requirement: Data Extraction Pre-conditions
The Loader Slice SHALL assume that the incoming JSON payload (produced by `extractData.pas`) follows a strict and correct formatting, incorporating fixes for index collision and response ordering.

#### Scenario: Parse extracted JSON correctly
- **WHEN** loading the JSON payload from the Pascal extraction script
- **THEN** LoaderSlice SHALL parse `stage_index` and `log_index` as separate fields to prevent collision
- **AND** SHALL parse explicit `order` (or `index`) fields for conversation responses to guarantee accurate tree traversal and ordering
