# Proposal: Refine Architecture Doc

## Goal
Refine the AIDD Architecture document to shift from a "Schema-First" (COBOL-style) approach to an "Interface-First" (Contracts/Header-style) approach.

## Context
The current `Schema-First_AIDD_Architecture_v2.md` advocates for "Atomic Context Injection" but relies on passing large Context structures, similar to COBOL copybooks. This causes tight coupling and "bucket brigade" parameter passing.
We identified that a "C-header style" approach—where modules strictly depend on abstract Interfaces (Contracts) defined in a separate layer—is superior for isolation and determinism.
The user also requested to remove specific Python implementation details (like `typing.Protocol`) from the high-level architecture doc to keep it language-neutral, and to drop the "COBOL vs C" metaphors in the final text in favor of standard terms like "Contract" and "Interface Definition".

## Capabilities

### `architecture-definition`
Defines the new "Interface-First" architecture principles and requirements for the documentation update.

## Impact
- **Renaming**: `Schema-First_AIDD_Architecture_v2.md` -> `Interface-First_AIDD_Architecture_v2.md` (or similar).
- **Content**: Major rewrite of the architecture philosophies and diagrams.
- **Workflow**: Future AI generation tasks will require Interface definitions as input instead of full Schemas.
