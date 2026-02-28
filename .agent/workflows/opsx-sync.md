---
description: Sync delta specs from a change to main specs
---

Sync delta specs from a change to main specs.

This is an **agent-driven** operation - you will read delta specs and directly edit main specs to apply the changes. This allows intelligent merging (e.g., adding a scenario without copying the entire requirement).

**Input**: Optionally specify a change name after `/opsx:sync` (e.g., `/opsx:sync add-auth`). If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show changes that have delta specs (under `specs/` directory).

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Find delta specs**

   Look for delta spec files in `openspec/changes/<name>/specs/*/spec.md`.

   Each delta spec file contains sections like:
   - `## ADDED Requirements` - New requirements to add
   - `## MODIFIED Requirements` - Changes to existing requirements
   - `## REMOVED Requirements` - Requirements to remove
   - `## RENAMED Requirements` - Requirements to rename (FROM:/TO: format)

   If no delta specs found, inform user and stop.

3. **For each delta spec, apply changes to main specs**

   For each capability with a delta spec at `openspec/changes/<name>/specs/<capability>/spec.md`:

   a. **First, determine the main specs path** by reading `openspec/config.yaml` and extracting the `specsPath` value (e.g., `specs/`, but it could be something else like `openspec/specs/`).

   b. **Read the delta spec** to understand the intended changes

   c. **Read the main spec** at `<specsPath>/<capability>/spec.md` (may not exist yet)
      - **IMPORTANT:** 既存のspecがある場合は、ファイルを作成し直すのではなく必ずその既存ファイルを開いて、そこに追加された要件や変更をマージすること。

   d. **Apply changes intelligently to the existing main spec**:
      **ADDED Requirements:**
      - If requirement doesn't exist in the existing main spec → add it
      - If requirement already exists the existing main spec → update it to match (treat as implicit MODIFIED)

      **MODIFIED Requirements:**
      - Find the requirement in the existing main spec
      - Apply the changes - this can be:
        - Adding new scenarios (don't need to copy existing ones)
        - Modifying existing scenarios
        - Changing the requirement description
      - Preserve scenarios/content not mentioned in the delta

      **REMOVED Requirements:**
      - Remove the entire requirement block from the main spec

      **RENAMED Requirements:**
      - Find the FROM requirement, rename to TO

   e. **Create new main spec** if capability doesn't exist yet:
      - Create `<specsPath>/<capability>/spec.md`
      - Add Purpose section (can be brief, mark as TBD)
      - Add Requirements section with the ADDED requirements

4. **Show summary**

   After applying all changes, summarize:
   - Which capabilities were updated
   - What changes were made (requirements added/modified/removed/renamed)

**Delta Spec Format Reference**

```markdown
## ADDED Requirements

### Requirement: New Feature
The system SHALL do something new.

#### Scenario: Basic case
- **WHEN** user does X
- **THEN** system does Y

## MODIFIED Requirements

### Requirement: Existing Feature
#### Scenario: New scenario to add
- **WHEN** user does A
- **THEN** system does B

## REMOVED Requirements

### Requirement: Deprecated Feature

## RENAMED Requirements

- FROM: `### Requirement: Old Name`
- TO: `### Requirement: New Name`
```

**Key Principle: Intelligent Merging**

Unlike programmatic merging, you can apply **partial updates**:
- To add a scenario, just include that scenario under MODIFIED - don't copy existing scenarios
- The delta represents *intent*, not a wholesale replacement
- Use your judgment to merge changes sensibly

**Output On Success**

```
## Specs Synced: <change-name>

Updated main specs:

**<capability-1>**:
- Added requirement: "New Feature"
- Modified requirement: "Existing Feature" (added 1 scenario)

**<capability-2>**:
- Created new spec file
- Added requirement: "Another Feature"

Main specs are now updated. The change remains active - archive when implementation is complete.
```

**Guardrails**
- Read both delta and main specs before making changes
- Preserve existing content not mentioned in delta
- If something is unclear, ask for clarification
- Show what you're changing as you go
- The operation should be idempotent - running twice should give same result