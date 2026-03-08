---
description: Start a new change using the experimental artifact workflow (OPSX)
---

Start a new change using the experimental artifact-driven approach.

## Local Customizations

This skill intentionally extends the baseline flow in `.agent/workflows/opsx-new.md`.

- If `openspec/review_template.md` exists, scaffold `openspec/changes/<name>/review.md` immediately after creating the change directory.
- Treat `review.md` as a change-specific review checklist, separate from the artifact graph.
- Preserve the review workflow discussed in this chat:
  - `openspec/review_standard.md` holds common review viewpoints
  - `openspec/review_template.md` defines the structure of change-specific review checklists
  - each new change should start with a writable `review.md` so verify can use it later


**Input**: The argument after `/opsx:new` is the change name (kebab-case), OR a description of what the user wants to build.

**Steps**

1. **If no input provided, ask what they want to build**

   Use the **AskUserQuestion tool** (open-ended, no preset options) to ask:
   > "What change do you want to work on? Describe what you want to build or fix."

   From their description, derive a kebab-case name (e.g., "add user authentication" → `add-user-auth`).

   **IMPORTANT**: Do NOT proceed without understanding what the user wants to build.

2. **Determine the workflow schema**

   Use the default schema (omit `--schema`) unless the user explicitly requests a different workflow.

   **Use a different schema only if the user mentions:**
   - A specific schema name → use `--schema <name>`
   - "show workflows" or "what workflows" → run `npx openspec schemas --json` and let them choose

   **Otherwise**: Omit `--schema` to use the default.

3. **Create the change directory**
   ```bash
   npx openspec new change "<name>"
   ```
   Add `--schema <name>` only if the user requested a specific workflow.
   This creates a scaffolded change at `openspec/changes/<name>/` with the selected schema.

4. **Create `review.md` from template when available**

   - Check whether `openspec/review_template.md` exists in the workspace root.
   - If it exists and `openspec/changes/<name>/review.md` does not exist yet:
     - Create `openspec/changes/<name>/review.md` by copying the template content.
   - Treat this file as a change-specific review checklist scaffold, not as an OpenSpec artifact node.
   - If the template does not exist, skip this step without failing.

5. **Show the artifact status**
   ```bash
   npx openspec status --change "<name>"
   ```
   This shows which artifacts need to be created and which are ready (dependencies satisfied).

6. **Get instructions for the first artifact**
   The first artifact depends on the schema. Check the status output to find the first artifact with status "ready".
   ```bash
   npx openspec instructions <first-artifact-id> --change "<name>"
   ```
   This outputs the template and context for creating the first artifact.

7. **STOP and wait for user direction**

**Output**

After completing the steps, summarize:
- Change name and location
- Schema/workflow being used and its artifact sequence
- Whether `openspec/changes/<name>/review.md` was scaffolded
- Current status (0/N artifacts complete)
- The template for the first artifact
- Prompt: "Ready to create the first artifact? Run `/opsx:continue` or just describe what this change is about and I'll draft it."

**Guardrails**
- Do NOT create any artifacts yet - just show the instructions
- Creating `review.md` from `openspec/review_template.md` is allowed and recommended; it is not treated as an artifact step
- Do NOT advance beyond showing the first artifact template
- If the name is invalid (not kebab-case), ask for a valid name
- If a change with that name already exists, suggest using `/opsx:continue` instead
- Pass --schema if using a non-default workflow

