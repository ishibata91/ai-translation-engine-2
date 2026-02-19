# Design: Architecture Doc Refinement

## Approach
This change focuses on refactoring the documentation itself. The core activity is rewriting `Schema-First_AIDD_Architecture_v2.md` to align with the new "Interface-First" philosophy.

## Implementation Steps

### 1. File Renaming
- Rename `Schema-First_AIDD_Architecture_v2.md` to `Interface-First_AIDD_Architecture_v2.md`.
- This signals the fundamental shift in philosophy.

### 2. Conceptual Refactoring
Rewrite the "Core Philosophy" section to emphasize:
- **Interfaces as Contracts**: Modules depend on abstract definitions, not implementation or data schemas.
- **Dependency Inversion**: High-level policy (Plans) and low-level detail (Impl) both depend on Abstractions (Interfaces).
- **Atomic Context**: Context injection is limited to likely-needed Interfaces, preventing "God Object" creation.

### 3. Terminology Updates
- **Schema** -> **Interface Definition** / **Contract**
- **COBOL/Copybook** -> **Header File** / **Abstraction Layer**
- **Sizing** -> Remains relevant but framed around "Complexity of Interface" rather than "Volume of Schema".

### 4. Diagrammatic Changes
Update Mermaid diagrams to visually represent:
- A central "Interface Layer".
- "Impl" nodes pointing TO the "Interface Layer".
- Removal of direct dependencies between "Impl" nodes.

## Risks
- **Confusion during transition**: Existing prompts or workflows might still reference "Schema-First".
- **Mitigation**: This doc update precedes the actual code/prompt refactoring (which will happen in subsequent changes), serving as the blueprint.

## Verification
- Review the generated markdown to ensure no "COBOL" or "Schema-First" vestiges remain.
- Check Mermaid diagrams for syntax errors and correct logic flow.
