# Data Loader
> Responsible for Phase 1 Data Foundation and domain models.

## Purpose
TBD: Parsing JSON files into domain models.

## Requirements

### Requirement: Structured domain models by context
The existing `models` package MUST retain the `ExtractedData` and related domain constructs, but split them from a single file into context-specific structural files (e.g. dialogue, quest, entity).

#### Scenario: Referencing domain models
- **WHEN** a component imports `github.com/ishibata91/ai-translation-engine-2/pkg/domain/models`
- **THEN** it can cleanly access models such as `NPC` or `Quest` without depending on a massive, monolithic file, while backward compatibility of the package path is maintained.
