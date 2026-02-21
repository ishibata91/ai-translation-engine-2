# Loader Slice Architecture
> Interface-First AIDD: Loader Vertical Slice

## Purpose
TBD: Provision of dependency injection for the loader module and encapsulation of internal data handling.

## Requirements

### Requirement: Loader slice DI Provider
The `pkg/loader` module MUST provide a dependency injection provider function via Google Wire, hiding its internal implementation details.

#### Scenario: DI Initialization
- **WHEN** the application starts and initializes its components
- **THEN** it resolves `contract.Loader` through the module's Wire provider without directly instantiating internal structs.

### Requirement: Internal process encapsulation
The sequential load and parallel decoding steps MUST be encapsulated within the `contract.Loader` implementation and not exposed to the Process Manager.

#### Scenario: Orchestrating the load process
- **WHEN** the Process Manager invokes `LoadExtractedJSON`
- **THEN** the slice internally coordinates file reading, decoding, and structuring, returning only the final `ExtractedData` or an error.
