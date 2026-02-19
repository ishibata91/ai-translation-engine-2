# Tasks: Architecture Doc Refinement

## 1. Preparation & Renaming

- [x] 1.1 Rename `Schema-First_AIDD_Architecture_v2.md` to `Interface-First_AIDD_Architecture_v2.md`
- [x] 1.2 Update the file title to "Interface-First AIDD (AI-Driven Development) Architecture"
- [x] 1.3 Update the Japanese subtitle to reflect "Interface-Driven" and "Dependency Separation"

## 2. Content Rewrite - Philosophy & Overview

- [x] 2.1 Rewrite "Executive Summary" to focus on Interface/Contracts instead of Schema/Data
- [x] 2.2 Rewrite "Core Philosophy" section
  - [x] Terminology: Replace "Atomic Context Injection" with "Interface-Limited Context Injection" (or similar)
  - [x] Terminology: Replace "Code is an Artifact" (keep concept but refine wording if needed)
  - [x] Concept: Define "Interface as the Contract" clearly
  - [x] Concept: Remove "COBOL" references and use standard engineering terms
- [x] 2.3 Rewrite "Architecture Overview" diagrams using Mermaid
  - [x] Visualise the dependency flow: Implementation -> Interface
  - [x] Remove "Schema" centric nodes

## 3. Content Rewrite - Technical Tactics

- [x] 3.1 Rewrite "Technical Tactics" section
  - [x] Remove specific Python `typing.Protocol` references (make language neutral)
  - [x] Explain "Protocol-Oriented Design" or "Contract-First Design" generally
  - [x] Explain "Mockability" benefits
- [x] 3.2 Update "Implementation Strategy" table
  - [x] Columns: Contracts, Logic, Wiring
  - [x] Define roles for Human vs AI in this new model

## 4. Verification

- [x] 4.1 Verify all text for "Schema-First" leftovers
- [x] 4.2 Validate Mermaid diagrams render correctly
- [x] 4.3 Ensure tone is professional and language-neutral
