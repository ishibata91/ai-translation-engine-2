# Tasks

## Shared Contracts
- `fix-review` の review feedback schema は `.codex/skills/fix-review/references/templates.md` を正本とし、field は `score / severity / location / violated_contract / required_delta / recheck / docs_sync_needed` に固定する。
- `fix-direction` / `fix-logging` / `log_instrumenter` の logging lifecycle は `add -> reproduce -> analyze -> remove` に固定し、cleanup は最終 accept 後かつ handoff 完了前に `fix-logging` の remove operation として実行する。
- 観測用一時ログの prefix は `fix-direction` / `fix-logging` / `log_instrumenter` の全境界で `[fix-trace]` に統一し、`[temp_fix-trace]` を残さない。
- `fix-logging` の handoff / add-remove 契約は `.codex/skills/fix-logging/references/templates.md` を正本とし、`SKILL.md` と agent 指示は同 template に従う。
- `investigation-direction` の handoff / conflict template は `.codex/skills/investigation-direction/references/templates.md` を正本とし、direction skill の返答形式を inline 記述へ戻さない。
- `investigation-distill` / `investigation-explorer` の探索 tool 記述は `server-filesystem` `go-llm-lens` `ts-lsp` に限定し、`find_by_name` `view_file` `grep` など現行規約外の tool 名を残さない。
- `impl-workplan` の `Section Plan` `Work Order` `tasks.md format` は `.codex/skills/impl-workplan/references/templates.md` を正本とし、`changes/<id>/tasks.md` の生成主体は `impl-workplan` だけに固定する。
- `impl-distill` の implementation packet schema は `.codex/skills/impl-distill/references/response-template.md` を正本とし、field は `task / scope / must_read / interfaces / entry_points / module_candidates / shared_contract_candidates / edit_boundary / validation_commands / constraints / acceptance / unknowns / handoff` に固定し、legacy `quality_gates` を残さない。
- `impl-direction` の正本 chain は `impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review` に固定し、section dispatch と reroute は `owner / depends_on / shared_contract / required_reading / validation_commands / acceptance` を固定後にだけ開始する。
- `aite2-ui-polish-orchestrate` は独立 user-facing 入口として扱わず、`impl-direction` 配下の specialized branch または internal helper としてのみ位置付ける。
- skill system の再検査は `.codex/skills/impl-workplan/scripts/validate-skill-contracts.ps1` を正本 command とし、missing template / deprecated tool name / logging prefix drift / workplan field mismatch を 1 回で検出できる状態にする。

## Phase Order
1. Hotfix
2. Workflow Normalization
3. Validation Automation

## Phase 1. Hotfix

### Dispatch Order
1. `fix-review-feedback-schema`
2. `fix-logging-lifecycle-alignment`
3. `investigation-direction-template-bootstrap`

## 1. Fix Review Feedback Schema

- section_id: `fix-review-feedback-schema`
- title: `Fix Review Feedback Schema`
- owner: backend
- goal: `fix-review` の schema 正本を `references/templates.md` に固定し、`SKILL.md` の出力契約と field parity を取る。
- depends_on: `[]`
- shared_contract: `fix-review` の review feedback field は `score / severity / location / violated_contract / required_delta / recheck / docs_sync_needed` で固定し、direction 側 loop 条件と不一致を残さない。
- owned_paths: `.codex/skills/fix-review/SKILL.md`, `.codex/skills/fix-review/references/templates.md`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/fix-review/SKILL.md`, `.codex/skills/fix-review/references/templates.md`
- validation_commands: `rg -n "score|severity|location|violated_contract|required_delta|recheck|docs_sync_needed" .codex/skills/fix-review/SKILL.md .codex/skills/fix-review/references/templates.md`; `rg -n "## review feedback" .codex/skills/fix-review/references/templates.md`
- acceptance: `references/templates.md` が唯一の schema 正本として成立している; `SKILL.md` の出力形式と loop 条件が template field と一致している; `score` 欠落や field 名の揺れが残っていない。
- [ ] 1.1 実装
- [ ] 1.2 検証

## 2. Fix Logging Lifecycle Alignment

- section_id: `fix-logging-lifecycle-alignment`
- title: `Fix Logging Lifecycle Alignment`
- owner: backend
- goal: `fix-direction` / `fix-logging` / `log_instrumenter` の cleanup 手順、add-remove 契約、prefix を一致させ、`fix-logging` が参照する `references/templates.md` を実在させる。
- depends_on: `[]`
- shared_contract: `fix-logging` は `add` と `remove` の両 operation を受け付け、`fix-direction` は最終 accept 後かつ完了 handoff 前に remove operation を起動する; 観測ログ prefix は `[fix-trace]` に固定し、agent は `log_additions` と `log_removals` を返せる。
- owned_paths: `.codex/skills/fix-direction/SKILL.md`, `.codex/skills/fix-direction/references/templates.md`, `.codex/skills/fix-logging/SKILL.md`, `.codex/skills/fix-logging/references/templates.md`, `.codex/agents/log_instrumenter.toml`
- forbidden_paths: `.codex/skills/fix-review/**`, `.codex/skills/fix-trace/**`, `.codex/skills/fix-work/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/fix-direction/SKILL.md`, `.codex/skills/fix-direction/references/templates.md`, `.codex/skills/fix-logging/SKILL.md`, `.codex/skills/fix-trace/references/templates.md`, `.codex/agents/log_instrumenter.toml`
- validation_commands: `Test-Path '.codex/skills/fix-logging/references/templates.md'`; `rg -n "\[fix-trace\]|log_additions|log_removals|cleanup|remove|add" .codex/skills/fix-direction/SKILL.md .codex/skills/fix-direction/references/templates.md .codex/skills/fix-logging/SKILL.md .codex/skills/fix-logging/references/templates.md .codex/agents/log_instrumenter.toml`; `rg -n "\[temp_fix-trace\]" .codex/skills/fix-direction/SKILL.md .codex/skills/fix-logging/SKILL.md .codex/skills/fix-logging/references/templates.md .codex/agents/log_instrumenter.toml`
- acceptance: `fix-direction` の cleanup 指示が `fix-logging` の remove contract と一致している; `.codex/skills/fix-logging/references/templates.md` が実在し add/remove packet 正本になっている; prefix は全境界で `[fix-trace]` に統一されている; `log_instrumenter` が add/remove 双方の戻り値契約を壊さない; `[temp_fix-trace]` が残っていない。
- [ ] 2.1 実装
- [ ] 2.2 検証

## 3. Investigation Direction Template Bootstrap

- section_id: `investigation-direction-template-bootstrap`
- title: `Investigation Direction Template Bootstrap`
- owner: backend
- goal: `investigation-direction` に `references/templates.md` を追加し、handoff / conflict template の正本を他 lane と同じ構造へ揃える。
- depends_on: `["fix-review-feedback-schema"]`
- shared_contract: direction skill の返答 schema は `references/templates.md` に置き、`SKILL.md` は契約参照と orchestration 手順だけを持つ。
- owned_paths: `.codex/skills/investigation-direction/SKILL.md`, `.codex/skills/investigation-direction/references/templates.md`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/investigation-direction/SKILL.md`, `.codex/skills/impl-direction/references/templates.md`
- validation_commands: `Test-Path '.codex/skills/investigation-direction/references/templates.md'`; `rg -n "references/templates.md|Conflict Handoff|Skill Invocation" .codex/skills/investigation-direction/SKILL.md .codex/skills/investigation-direction/references/templates.md`
- acceptance: `.codex/skills/investigation-direction/references/templates.md` が新設または正本化されている; `SKILL.md` が handoff / conflict 出力を template 参照へ委譲している; inline-only schema が残っていない。
- [ ] 3.1 実装
- [ ] 3.2 検証

## Phase 2. Workflow Normalization

### Dispatch Order
4. `investigation-tool-contract-normalization`
5. `impl-workplan-schema-normalization`
6. `impl-direction-chain-normalization`
7. `entry-routing-and-agent-boundary`

## 4. Investigation Tool Contract Normalization

- section_id: `investigation-tool-contract-normalization`
- title: `Investigation Tool Contract Normalization`
- owner: backend
- goal: `investigation-distill` / `investigation-explorer` から現行規約外の tool 名を除去し、探索導線を `server-filesystem` `go-llm-lens` `ts-lsp` に揃える。
- depends_on: `["investigation-direction-template-bootstrap"]`
- shared_contract: investigation lane の tool 記述は `server-filesystem` による検索・読取、`pkg/` 以下では `go-llm-lens`、`frontend/src/` 以下では `ts-lsp` を正本とし、旧 tool alias や汎用 `grep` 指示を残さない。
- owned_paths: `.codex/skills/investigation-distill/SKILL.md`, `.codex/skills/investigation-explorer/SKILL.md`
- forbidden_paths: `AGENTS.md`, `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `AGENTS.md`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/investigation-distill/SKILL.md`, `.codex/skills/investigation-explorer/SKILL.md`
- validation_commands: `rg -n "server-filesystem|go-llm-lens|ts-lsp" .codex/skills/investigation-distill/SKILL.md .codex/skills/investigation-explorer/SKILL.md`; `rg -n "find_by_name|view_file|grep" .codex/skills/investigation-distill/SKILL.md .codex/skills/investigation-explorer/SKILL.md`
- acceptance: `investigation-distill` と `investigation-explorer` が AGENTS.md の MCP 利用規約を参照している; `find_by_name` `view_file` `grep` など規約外の tool 名が残っていない; 探索対象ごとの正本 tool が文章上で判別できる。
- [ ] 4.1 実装
- [ ] 4.2 検証

## 5. Impl Planning Contract Normalization

- section_id: `impl-workplan-schema-normalization`
- title: `Impl Planning Contract Normalization`
- owner: backend
- goal: `impl-distill -> impl-workplan` の implementation packet / section schema を `validation_commands` 正本へ揃え、legacy field を除去したうえで `tasks.md` 生成責務を `impl-workplan` に限定する。
- depends_on: `["fix-review-feedback-schema", "investigation-direction-template-bootstrap"]`
- shared_contract: `impl-distill -> impl-workplan` の packet / section contract は `validation_commands` を唯一の validation field とし、`quality_gates` や未固定 validation field を worker に流さない。
- owned_paths: `.codex/skills/impl-distill/SKILL.md`, `.codex/skills/impl-distill/references/response-template.md`, `.codex/skills/impl-workplan/SKILL.md`, `.codex/skills/impl-workplan/references/templates.md`, `.codex/skills/impl-workplan/references/response-template.md`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/impl-distill/SKILL.md`, `.codex/skills/impl-distill/references/response-template.md`, `.codex/skills/impl-workplan/SKILL.md`, `.codex/skills/impl-workplan/references/templates.md`, `.codex/skills/impl-workplan/references/response-template.md`, `.codex/skills/impl-direction/references/templates.md`
- validation_commands: `rg -n "quality_gates|validation_commands" .codex/skills/impl-distill/SKILL.md .codex/skills/impl-distill/references/response-template.md`; `rg -n "Section Plan|Work Order|tasks.md format|owner: frontend \| backend|validation_commands" .codex/skills/impl-workplan/references/templates.md`; `rg -n "implementation packet|tasks.md.*impl-workplan|section plan|work order|quality_gates|validation_commands" .codex/skills/impl-distill/SKILL.md .codex/skills/impl-distill/references/response-template.md .codex/skills/impl-workplan/SKILL.md .codex/skills/impl-workplan/references/response-template.md`
- acceptance: `impl-distill` の packet 契約と response template が `validation_commands` field に揃い、legacy `quality_gates` が残っていない; `impl-workplan` の `references/templates.md` と response / skill docs が packet consumer として相互に矛盾していない; `tasks.md` の生成主体が `impl-workplan` 以外へ戻っていない; section planning に owner 未確定や shared contract 未固定を許す文言が残っていない。
- [ ] 5.1 実装
- [ ] 5.2 検証

## 6. Impl Direction Chain Normalization

- section_id: `impl-direction-chain-normalization`
- title: `Impl Direction Chain Normalization`
- owner: backend
- goal: `impl-direction` の正本 chain と Workplan Summary / Section Dispatch / Review Reroute schema を `impl-workplan` 前提へ一本化し、section dispatch と review reroute を full section contract に揃える。
- depends_on: `["impl-workplan-schema-normalization"]`
- shared_contract: `impl-direction` は `impl-distill` と `impl-workplan` の packet を待つ orchestration-only controller とし、受領 / 差し戻し時の section schema では `shared_contract / required_reading / validation_commands / acceptance` を省略しない。
- owned_paths: `.codex/skills/impl-direction/SKILL.md`, `.codex/skills/impl-direction/references/templates.md`, `.codex/skills/impl-direction/references/quality-checklist.md`, `.codex/skills/impl-direction/references/examples.md`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-workplan/**`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/impl-direction/SKILL.md`, `.codex/skills/impl-direction/references/templates.md`, `.codex/skills/impl-direction/references/quality-checklist.md`, `.codex/skills/impl-workplan/SKILL.md`, `.codex/skills/impl-workplan/references/templates.md`
- validation_commands: `rg -n "impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review|workplan_builder|tasks.md: optional" .codex/skills/impl-direction/SKILL.md .codex/skills/impl-direction/references/templates.md .codex/skills/impl-direction/references/quality-checklist.md`; `rg -n "Workplan Summary|Section Dispatch|Review Reroute|affected_sections|shared_contract|required_reading|validation_commands|acceptance" .codex/skills/impl-direction/references/templates.md`
- acceptance: `impl-direction` の chain が `impl-workplan` 前提に一本化されている; `Workplan Summary` の section schema が `impl-workplan` の full contract を保持し validator が `validation_commands` 欠落を再検出しない; routing matrix が `ui.md` 不在だけで backend-only と判定しない; section dispatch と review reroute が template 契約と一致している。
- [ ] 6.1 実装
- [ ] 6.2 検証

## 7. Entry Routing And Agent Boundary

- section_id: `entry-routing-and-agent-boundary`
- title: `Entry Routing And Agent Boundary`
- owner: backend
- goal: `aite2-ui-polish-orchestrate` を独立 user-facing 入口から外し、`workplan_builder` agent を `impl-workplan` 専用境界として同期する。
- depends_on: `["impl-direction-chain-normalization"]`
- shared_contract: specialized skill は direction skill を上書きせず、`workplan_builder` agent は `impl-workplan` 以外の実装責務を持たない。
- owned_paths: `.codex/skills/aite2-ui-polish-orchestrate/SKILL.md`, `.codex/skills/aite2-ui-polish-orchestrate/references/**`, `.codex/agents/workplan_builder.toml`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/**`, `.codex/agents/log_instrumenter.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/aite2-ui-polish-orchestrate/SKILL.md`, `.codex/skills/impl-direction/SKILL.md`, `.codex/agents/workplan_builder.toml`
- validation_commands: `rg -n "user-facing|入口|impl-direction|internal helper|specialized branch" .codex/skills/aite2-ui-polish-orchestrate/SKILL.md .codex/skills/aite2-ui-polish-orchestrate/references`; `rg -n "workplan_builder|impl-workplan|changes/<id>/tasks.md|section-based workplan" .codex/agents/workplan_builder.toml`
- acceptance: `aite2-ui-polish-orchestrate` が独立 user-facing 入口として読めない; `impl-direction` 配下の補助フローまたは internal helper として位置付けが明示されている; `workplan_builder.toml` が `impl-workplan` 専用の section planning / tasks 生成境界だけを記述している。
- [ ] 7.1 実装
- [ ] 7.2 検証

## Phase 3. Validation Automation

### Dispatch Order
8. `skill-contract-validation-automation`

## 8. Skill Contract Validation Automation

- section_id: `skill-contract-validation-automation`
- title: `Skill Contract Validation Automation`
- owner: backend
- goal: overhaul 対象の shared contract を反復検査できる validation command と checklist を追加し、hotfix / normalization 後に同じ監査を再実行できるようにする。
- depends_on: `["fix-logging-lifecycle-alignment", "investigation-tool-contract-normalization", "entry-routing-and-agent-boundary"]`
- shared_contract: validation automation は missing reference template、deprecated tool name、logging prefix drift、`validation_commands` を含む workplan field mismatch を 1 つの正本 command で検出し、review 前の必須 gate とする。
- owned_paths: `.codex/skills/impl-workplan/scripts/validate-skill-contracts.ps1`, `.codex/skills/impl-workplan/references/validation-checklist.md`
- forbidden_paths: `.codex/skills/fix-direction/**`, `.codex/skills/fix-logging/**`, `.codex/skills/fix-review/**`, `.codex/skills/investigation-direction/**`, `.codex/skills/investigation-distill/**`, `.codex/skills/investigation-explorer/**`, `.codex/skills/impl-direction/**`, `.codex/skills/impl-workplan/SKILL.md`, `.codex/skills/impl-workplan/references/templates.md`, `.codex/skills/aite2-ui-polish-orchestrate/**`, `.codex/agents/log_instrumenter.toml`, `.codex/agents/workplan_builder.toml`, `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `changes/skill-system-overhaul/tasks.md`
- required_reading: `changes/skill-system-overhaul/review.md`, `changes/skill-system-overhaul/scenarios.md`, `changes/skill-system-overhaul/logic.md`, `.codex/skills/fix-direction/SKILL.md`, `.codex/skills/fix-logging/SKILL.md`, `.codex/skills/investigation-distill/SKILL.md`, `.codex/skills/investigation-explorer/SKILL.md`, `.codex/skills/impl-direction/SKILL.md`, `.codex/skills/impl-workplan/references/templates.md`, `.codex/agents/log_instrumenter.toml`
- validation_commands: `Test-Path '.codex/skills/impl-workplan/scripts/validate-skill-contracts.ps1'`; `Test-Path '.codex/skills/impl-workplan/references/validation-checklist.md'`; `powershell -ExecutionPolicy Bypass -File .codex/skills/impl-workplan/scripts/validate-skill-contracts.ps1`
- acceptance: validation command が 1 回で required delta 対象の崩れを検出できる; checklist が Hotfix / Workflow Normalization / Validation Automation の各 phase で何を確認するかを明示している; future skill 追加時も同 command を再利用できる。
- [ ] 8.1 実装
- [ ] 8.2 検証
