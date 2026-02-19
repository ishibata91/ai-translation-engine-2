# Spec: Architecture Definition

## Purpose
Define the structure and principles of the new "Interface-First" AIDD Architecture.

## ADDED Requirements

### Requirement: Architecture Document Renaming
The architecture document SHALL be renamed to reflect the new philosophy.

#### Scenario: Rename File
- **WHEN** the update is applied
- **THEN** the file `Schema-First_AIDD_Architecture_v2.md` is renamed to `Interface-First_AIDD_Architecture_v2.md`
- **AND** references to "Schema-First" in the content are replaced with "Interface-First"

### Requirement: Interface-First Philosophy
The architecture SHALL advocate for dependency on abstract Interfaces/Contracts rather than data Schemas.

#### Scenario: Dependency Rule
- **WHEN** defining module dependencies
- **THEN** modules MUST depend only on "Interface Definitions" (Header/Contract)
- **AND** modules MUST NOT depend on concrete implementations or large context structures

#### Scenario: Atomic Context Injection
- **WHEN** injecting context into an AI generation task
- **THEN** the context SHALL contain only the Implementation Plan and the relevant Interface Definitions
- **AND** it SHALL NOT contain the full project source tree or implementation details

### Requirement: Language Neutrality
The architecture document SHALL be language-neutral and avoid specific implementation details.

#### Scenario: Remove Python Specifics
- **WHEN** describing technical tactics
- **THEN** mentions of `typing.Protocol`, `ABC`, or specific Python syntax are removed or generalized
- **AND** concepts are described using universal terms like "Interface", "Contract", or "Protocol"

### Requirement: Terminology Update
The architecture document SHALL use professional engineering terminology suitable for the new design.

#### Scenario: Remove COBOL Metaphors
- **WHEN** explaining the rationale
- **THEN** references to "COBOL", "Copybooks", or "C-Header style" are removed from the normative text (though "header-style" can be used as an analogy if helpful, standard terms are preferred)
- **AND** focus is placed on "Decoupling", "Contracts", and "Determinism"

### Requirement: Diagram Updates
The architecture diagrams SHALL reflect the Interface-First flow.

#### Scenario: Update Dependency Graph
- **WHEN** visualizing the architecture
- **THEN** the diagram shows "Interfaces" as the central dependency
- **AND** "Implementations" depend on "Interfaces"
- **AND** "Implementations" do not cross-reference each other directly
