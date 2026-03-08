---
description: Verify implementation matches change artifacts before archiving
---

Verify that an implementation matches the change artifacts (specs, tasks, design, review checklists).

## Local Customizations

This skill intentionally extends the baseline flow in `.agent/workflows/opsx-verify.md`.

- Read `openspec/review_standard.md` as the global review viewpoints catalog.
- Read `openspec/review_template.md` as the reference structure for change-specific review checklists.
- Read `openspec/changes/<name>/review.md` when present.
- Use a diff-first review flow: inspect `git diff` first, then search wider only if needed.
- Preserve the project-specific review viewpoints discussed in this chat, especially:
  - backend-specific review points based on `backend_coding_standards.md`
  - frontend-specific review points based on `frontend_architecture.md` and `frontend_coding_standards.md`
  - telemetry/logging review points centered on `pkg/infrastructure/telemetry`

**Input**: Optionally specify a change name after `/opsx:verify` (e.g., `/opsx:verify add-auth`). If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `npx openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show changes that have implementation tasks (tasks artifact exists).
   Include the schema used for each change if available.
   Mark changes with incomplete tasks as "(In Progress)".

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Check status to understand the schema**
   ```bash
   npx openspec status --change "<name>" --json
   ```
   Parse the JSON to understand:
   - `schemaName`: The workflow being used (e.g., "spec-driven")
   - Which artifacts exist for this change

3. **Get the change directory and load artifacts**

   ```bash
   npx openspec instructions apply --change "<name>" --json
   ```

   This returns the change directory and context files. Read all available artifacts from `contextFiles`.

3.5 **Load review standards**

   - Read `openspec/review_standard.md` from the workspace root if it exists.
   - Read `openspec/review_template.md` if it exists, to understand the expected structure of change-specific review checklists.
   - Read `openspec/changes/<name>/review.md` if it exists.
   - Treat `openspec/review_standard.md` as the global review viewpoints catalog.
   - Treat `openspec/review_template.md` as the reference template for how a change-specific `review.md` should be organized.
   - Treat `openspec/changes/<name>/review.md` as the change-specific review checklist.
   - If `openspec/review_standard.md` is missing, add a WARNING noting that the global review standard is missing.
   - If `openspec/changes/<name>/review.md` is missing, do not fail verification automatically, but note that no change-specific review checklist was provided.

3.6 **Identify changed files first**

   - Before broad codebase exploration, inspect the git diff and identify the changed files first.
   - Start with commands such as:
     ```bash
     git diff --name-only --relative HEAD
     git diff --stat --relative HEAD
     ```
   - Use the diff as the primary search boundary for verification.
   - Focus requirement/design/review checks on changed files and directly related packages first.
   - Only broaden to wider codebase search when the diff does not provide enough evidence to verify a requirement or checklist item.

4. **Initialize verification report structure**

   Create a report structure with three dimensions:
   - **Completeness**: Track tasks and spec coverage
   - **Correctness**: Track requirement implementation and scenario coverage
   - **Coherence**: Track design adherence, pattern consistency, and review checklist conformance

   Each dimension can have CRITICAL, WARNING, or SUGGESTION issues.

5. **Verify Completeness**

   **Task Completion**:
   - If tasks.md exists in contextFiles, read it
   - Parse checkboxes: `- [ ]` (incomplete) vs `- [x]` (complete)
   - Count complete vs total tasks
   - If incomplete tasks exist:
     - Add CRITICAL issue for each incomplete task
     - Recommendation: "Complete task: <description>" or "Mark as done if already implemented"

   **Spec Coverage**:
   - If delta specs exist in `openspec/changes/<name>/specs/`:
     - Extract all requirements (marked with "### Requirement:")
     - For each requirement:
       - Search codebase for keywords related to the requirement
       - Assess if implementation likely exists
     - If requirements appear unimplemented:
       - Add CRITICAL issue: "Requirement not found: <requirement name>"
       - Recommendation: "Implement requirement X: <description>"

6. **Verify Correctness**

   **Requirement Implementation Mapping**:
   - For each requirement from delta specs:
     - Search changed files first for implementation evidence
     - If needed, expand to directly related packages
     - If found, note file paths and line ranges
     - Assess if implementation matches requirement intent
     - If divergence detected:
       - Add WARNING: "Implementation may diverge from spec: <details>"
       - Recommendation: "Review <file>:<lines> against requirement X"

   **Scenario Coverage**:
   - For each scenario in delta specs (marked with "#### Scenario:"):
     - Check changed files first to see if the conditions are handled in code
     - Check if tests exist covering the scenario
     - If scenario appears uncovered:
       - Add WARNING: "Scenario not covered: <scenario name>"
       - Recommendation: "Add test or implementation for scenario: <description>"

7. **Verify Coherence**

   **Design Adherence**:
   - If design.md exists in contextFiles:
     - Extract key decisions (look for sections like "Decision:", "Approach:", "Architecture:")
     - Verify implementation follows those decisions
     - If contradiction detected:
       - Add WARNING: "Design decision not followed: <decision>"
       - Recommendation: "Update implementation or revise design.md to match reality"
   - If no design.md: Skip design adherence check, note "No design.md to verify against"

   **Code Pattern Consistency**:
   - Review new code for consistency with project patterns
   - Check file naming, directory structure, coding style
   - If significant deviations found:
     - Add SUGGESTION: "Code pattern deviation: <details>"
     - Recommendation: "Consider following project pattern: <example>"

   **Review Checklist Conformance**:
   - If `openspec/review_standard.md` exists:
     - Extract the review viewpoints and apply them during verification.
     - Include findings when the implementation diverges from those viewpoints.
   - If `openspec/changes/<name>/review.md` exists:
     - Check each item in the change-specific review checklist against the implementation.
     - If a checklist item is not satisfied:
       - Add WARNING: "Review checklist item not satisfied: <item>"
       - Recommendation: "Review implementation against change review checklist item: <item>"
   - If neither review file exists:
     - Note that review checklist verification was skipped due to missing review artifacts.

8. **Generate Verification Report**

   **Summary Scorecard**:
   ```
   ## Verification Report: <change-name>

   ### Summary
   | Dimension    | Status           |
   |--------------|------------------|
   | Completeness | X/Y tasks, N reqs|
   | Correctness  | M/N reqs covered |
   | Coherence    | Followed/Issues  |
   ```

   **Issues by Priority**:

   1. **CRITICAL** (Must fix before archive):
      - Incomplete tasks
      - Missing requirement implementations
      - Each with specific, actionable recommendation

   2. **WARNING** (Should fix):
      - Spec/design divergences
      - Missing scenario coverage
      - Review checklist mismatches
      - Each with specific recommendation

   3. **SUGGESTION** (Nice to fix):
      - Pattern inconsistencies
      - Minor improvements
      - Each with specific recommendation

   **Final Assessment**:
   - If CRITICAL issues: "X critical issue(s) found. Fix before archiving."
   - If only warnings: "No critical issues. Y warning(s) to consider. Ready for archive (with noted improvements)."
   - If all clear: "All checks passed. Ready for archive."

**Verification Heuristics**

- **Completeness**: Focus on objective checklist items (checkboxes, requirements list)
- **Correctness**: Use keyword search, file path analysis, reasonable inference - don't require perfect certainty
- **Diff-First Review**: Prefer reviewing git-changed files and their immediate neighbors before exploring the whole repository
- **Coherence**: Look for glaring inconsistencies, don't nitpick style
- **False Positives**: When uncertain, prefer SUGGESTION over WARNING, WARNING over CRITICAL
- **Actionability**: Every issue must have a specific recommendation with file/line references where applicable

**Graceful Degradation**

- If only tasks.md exists: verify task completion only, skip spec/design checks
- If tasks + specs exist: verify completeness and correctness, skip design
- If full artifacts: verify all three dimensions
- Review checklist verification should degrade gracefully when `openspec/review_standard.md` or `review.md` is absent
- Always note which checks were skipped and why

**Output Format**

Use clear markdown with:
- Table for summary scorecard
- Grouped lists for issues (CRITICAL/WARNING/SUGGESTION)
- Code references in format: `file.ts:123`
- Specific, actionable recommendations
- No vague suggestions like "consider reviewing"
