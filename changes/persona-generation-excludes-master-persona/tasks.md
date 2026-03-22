## 1. Translation Input Persona Projection

- section_id: backend-translationinput-persona-projection
- owner: backend
- goal: `translationinput` artifact から persona phase 用の候補投影を提供し、`source_plugin + speaker_id` 正規化に必要な NPC / dialogue 情報を workflow へ渡せるようにする。
- depends_on: []
- shared_contract: [`translationinput.PersonaInput`, `translationinput.PersonaNPC`, `translationinput.PersonaDialogue`, `translationinput.Repository.LoadPersonaInput(ctx, taskID string) (PersonaInput, error)`]
- owned_paths: [`pkg/artifact/translationinput/contract.go`, `pkg/artifact/translationinput/repository.go`, `pkg/artifact/translationinput/repository_test.go`]
- forbidden_paths: [`pkg/slice/translationflow/*`, `pkg/workflow/*`, `pkg/controller/*`, `frontend/src/*`, `pkg/artifact/master_persona_artifact/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `changes/persona-generation-excludes-master-persona/logic.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `pkg/artifact/translationinput/contract.go`, `pkg/artifact/translationinput/repository.go`, `pkg/artifact/master_persona_artifact/contract.go`, `pkg/artifact/master_persona_artifact/repository.go`]
- validation_commands: [`go test ./pkg/artifact/translationinput`]
- acceptance: [`persona 候補投影が NPC / dialogue / source hint / source plugin を返せる`, `既存 terminology 投影を流用せず persona 専用 API を追加する`, `lookup key 補完に必要な `UNKNOWN` / file hint フォールバックの材料を DTO 側で失わない`]
- [x] 1.1 実装
- [x] 1.2 検証

## 2. Translation Flow Persona Read Contracts

- section_id: backend-translationflow-persona-store-contracts
- owner: backend
- goal: workflow が `translationinput` / `master_persona_artifact` の詳細型へ直接依存しないよう、persona 候補・既存 final lookup を `translationflow` slice ローカル contract に閉じ込める。
- depends_on: [`backend-translationinput-persona-projection`]
- shared_contract: [`translationflow.PersonaCandidateInput`, `translationflow.PersonaCandidate`, `translationflow.PersonaDialogueExcerpt`, `translationflow.PersonaFinalSummary`, `translationflow.PersonaLookupKey`, `translationflow.Service.LoadPersonaCandidates(ctx, taskID string) (PersonaCandidateInput, error)`, `translationflow.Service.FindPersonaFinal(ctx, key PersonaLookupKey) (PersonaFinalSummary, bool, error)`]
- owned_paths: [`pkg/slice/translationflow/contract.go`, `pkg/slice/translationflow/service.go`]
- forbidden_paths: [`pkg/artifact/translationinput/*`, `pkg/workflow/*`, `pkg/controller/*`, `frontend/src/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/logic.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `docs/governance/architecture/spec.md`, `pkg/slice/translationflow/contract.go`, `pkg/slice/translationflow/service.go`, `pkg/artifact/translationinput/contract.go`, `pkg/artifact/translationinput/repository.go`, `pkg/artifact/master_persona_artifact/contract.go`, `pkg/artifact/master_persona_artifact/repository.go`]
- validation_commands: [`go test ./pkg/slice/translationflow ./pkg/artifact/translationinput ./pkg/artifact/master_persona_artifact`]
- acceptance: [`workflow が `translationinput.PersonaInput` と `master_persona_artifact.LookupKey` を直接 import せずに候補・既存 final を扱える`, `translationflow slice が artifact 依存を吸収しつつ orchestration ルールは持ち込まない`, `既存 load / terminology 用 API の挙動を壊さない`]
- [ ] 2.1 実装
- [ ] 2.2 検証

## 3. Master Persona Execution Contract

- section_id: backend-master-persona-execution-contract
- owner: backend
- goal: persona phase の bootstrap/resume/runtime snapshot を `translation_project.task_id` 境界で扱う workflow-local contract を定義し、`request/prompt` を実実装へ伝搬させる。
- depends_on: []
- shared_contract: [`workflow.PersonaExecutionInput`, `workflow.PersonaRuntimeEntry`, `workflow.MasterPersona.RunPersonaPhase(ctx, input PersonaExecutionInput) error`, `workflow.MasterPersona.ListPersonaRuntime(ctx, taskID string) ([]PersonaRuntimeEntry, error)`]
- owned_paths: [`pkg/workflow/master_persona.go`, `pkg/workflow/master_persona_service.go`, `pkg/workflow/master_persona_service_test.go`]
- forbidden_paths: [`pkg/controller/*`, `frontend/src/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/scenarios.md`, `changes/persona-generation-excludes-master-persona/logic.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `docs/governance/architecture/spec.md`, `pkg/workflow/master_persona.go`, `pkg/workflow/master_persona_service.go`, `pkg/workflow/task/manager.go`, `pkg/runtime/queue/queue.go`]
- validation_commands: [`go test ./pkg/workflow ./pkg/runtime/queue`]
- acceptance: [`別 task を新規発行せず `translation_project.task_id` 配下で初回 bootstrap と resume が成立する`, `request/prompt` が persona request 生成契約へ実際に渡る`, `workflow が `runtimequeue.JobRequest` へ直接依存せず runtime snapshot を受け取れる`]
- [ ] 3.1 実装
- [ ] 3.2 検証

## 4. Translation Flow Persona Workflow

- section_id: backend-translationflow-persona-workflow
- owner: backend
- goal: persona phase の preview / execute / resume / retry / no-op 完了 / partial failure 復元を、slice と master persona の local contract だけで orchestration する。
- depends_on: [`backend-translationflow-persona-store-contracts`, `backend-master-persona-execution-contract`]
- shared_contract: [`workflow.PersonaTargetPreviewRow`, `workflow.PersonaDialogueView`, `workflow.PersonaTargetPreviewPage`, `workflow.RunTranslationFlowPersonaPhaseInput`, `workflow.PersonaPhaseResult`, `workflow.TranslationFlow.ListTranslationFlowPersonaTargets`, `workflow.TranslationFlow.RunTranslationFlowPersonaPhase`, `workflow.TranslationFlow.GetTranslationFlowPersonaPhase`]
- owned_paths: [`pkg/workflow/translation_flow.go`, `pkg/workflow/translation_flow_service.go`, `pkg/workflow/translation_flow_service_test.go`]
- forbidden_paths: [`pkg/controller/*`, `frontend/src/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `changes/persona-generation-excludes-master-persona/logic.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `docs/governance/architecture/spec.md`, `pkg/workflow/translation_flow.go`, `pkg/workflow/translation_flow_service.go`, `pkg/workflow/translation_flow_service_test.go`, `pkg/workflow/master_persona.go`, `pkg/slice/translationflow/contract.go`]
- validation_commands: [`go test ./pkg/workflow ./pkg/controller ./pkg/runtime/queue`]
- acceptance: [`fresh translation task で request 未生成なら bootstrap し、生成済みなら resume する`, `preview と execute が同一 planner を使い、既存 final は request 化しない`, `pending 0 件では runtime を呼ばず no-op 完了になる`, `queue 未生成でも `empty` / `ready` / `cachedOnly` を正常導出できる`, `workflow が artifact/runtime の詳細型へ直接依存しない`]
- [ ] 4.1 実装
- [ ] 4.2 検証

## 5. Translation Flow Persona Runtime Wiring

- section_id: backend-translationflow-persona-runtime-wiring
- owner: backend
- goal: `main.go` で revision 後の `translationflow` / `master persona` / `translation flow` contract を注入し、既存 app 起動配線を壊さずに runtime wiring を完成させる。
- depends_on: [`backend-master-persona-execution-contract`, `backend-translationflow-persona-workflow`]
- shared_contract: [`translationflow.NewService(...)`, `workflow.NewMasterPersonaService(...)`, `workflow.NewTranslationFlowService(parser, store, terminology, personaWorkflow, executor, notifier)`]
- owned_paths: [`main.go`]
- forbidden_paths: [`pkg/artifact/translationinput/*`, `pkg/slice/translationflow/*`, `pkg/workflow/*`, `pkg/controller/*`, `frontend/src/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/logic.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `main.go`, `pkg/workflow/master_persona.go`, `pkg/workflow/master_persona_service.go`, `pkg/workflow/translation_flow_service.go`, `pkg/slice/translationflow/service.go`]
- validation_commands: [`go test . ./pkg/workflow ./pkg/runtime/queue`]
- acceptance: [`translation flow workflow が revision 後の persona contract を受け取る`, `translationflow slice / master persona workflow / controller 既存 wiring を壊さない`, `runtime wiring が compile 順と依存順の両方で整合する`]
- [ ] 5.1 実装
- [ ] 5.2 検証

## 6. Task Controller Persona API

- section_id: backend-task-controller-persona-api
- owner: backend
- goal: Wails 向け `TaskController` の persona phase API を revision 後の workflow 契約に追従させ、`translation_project.task_id` 境界のまま preview / run / get を公開する。
- depends_on: [`backend-translationflow-persona-workflow`]
- shared_contract: [`TaskController.ListTranslationFlowPersonaTargets`, `TaskController.RunTranslationFlowPersona`, `TaskController.GetTranslationFlowPersona`, `workflow.RunTranslationFlowPersonaPhaseInput`, `workflow.PersonaPhaseResult`]
- owned_paths: [`pkg/controller/task_controller.go`, `pkg/controller/task_controller_test.go`]
- forbidden_paths: [`pkg/artifact/translationinput/*`, `pkg/slice/translationflow/*`, `frontend/src/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/scenarios.md`, `docs/workflow/translation-flow-persona-phase/spec.md`, `pkg/controller/task_controller.go`, `pkg/controller/task_controller_test.go`, `pkg/workflow/translation_flow.go`]
- validation_commands: [`go test ./pkg/controller ./pkg/workflow`]
- acceptance: [`controller が別 persona task 前提を持たず、同一 `translation_project.task_id` のまま persona phase を呼び出す`, `preview / run / get の 3 API が workflow revision 契約に沿って公開される`, `既存 load / terminology API の解決フローを壊さない`]
- [ ] 6.1 実装
- [ ] 6.2 検証

## 7. Translation Flow Persona Types And Adapters

- section_id: frontend-translationflow-persona-types-adapters
- owner: frontend
- goal: backend persona DTO を受ける TypeScript の state / payload / adapter 契約を追加し、persona state machine を表現できるようにする。
- depends_on: [`backend-task-controller-persona-api`]
- shared_contract: [`PersonaTargetPreviewRow`, `PersonaDialogueView`, `PersonaTargetPreviewPage`, `PersonaPhaseSummary`, `PersonaTargetViewState`, `WailsPersonaTargetPreviewPage`, `WailsPersonaPhaseResult`]
- owned_paths: [`frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/translationFlow/adapters.ts`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/components/translation-flow/*`, `frontend/src/pages/TranslationFlow.tsx`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `docs/frontend/translation-flow-persona-ui/spec.md`, `frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/translationFlow/adapters.ts`, `pkg/workflow/translation_flow.go`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/hooks/features/translationFlow/types.ts src/hooks/features/translationFlow/adapters.ts`]
- acceptance: [`types が loadingTargets / empty / ready / cachedOnly / running / completed / partialFailed / failed を表現できる`, `adapter が snake_case / camelCase の persona payload を正規化する`, `terminology 既存型を壊さず persona 用型を追加する`]
- [x] 7.1 実装
- [x] 7.2 検証

## 8. Translation Flow Persona Hook

- section_id: frontend-translationflow-persona-hook
- owner: frontend
- goal: `useTranslationFlow` を revision 後の backend status / run 契約に追従させ、初回 persona 表示が `failed` に倒れないことを regression test で固定する。
- depends_on: [`backend-task-controller-persona-api`, `frontend-translationflow-persona-types-adapters`]
- shared_contract: [`useTranslationFlow` persona state / actions contract, `TaskController` persona bindings, `PersonaTargetPreviewPage`, `PersonaPhaseSummary`]
- owned_paths: [`frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/translationFlow/adapters.ts`, `frontend/src/components/translation-flow/*`, `frontend/src/pages/TranslationFlow.tsx`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `docs/frontend/translation-flow-persona-ui/spec.md`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`, `frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/translationFlow/adapters.ts`]
- validation_commands: [`cd frontend && npm run test -- src/hooks/features/translationFlow/useTranslationFlow.test.tsx`, `cd frontend && npm run typecheck`]
- acceptance: [`hook が persona phase を terminology の次・summary の前で制御する`, `mount/tab-change 時に queue 未生成でも `failed` へ落ちない`, `cachedOnly / empty / running / partialFailed の操作可否を state で制御する`, `persona phase の config / prompt は revision 後の backend 契約と一致する`]
- [ ] 8.1 実装
- [ ] 8.2 検証

## 9. Persona Panel Rendering

- section_id: frontend-persona-panel-rendering
- owner: frontend
- goal: revision 後の state semantics に合わせて `PersonaPanel` の summary/list/detail/footer を整え、初回表示と partialFailed 表示を安定させる。
- depends_on: [`frontend-translationflow-persona-types-adapters`, `frontend-translationflow-persona-hook`]
- shared_contract: [`PersonaPanelProps`, `PersonaTargetPreviewRow`, `PersonaPhaseSummary`, `PersonaDialogueView`, `PersonaTargetViewState`]
- owned_paths: [`frontend/src/components/translation-flow/PersonaPanel.tsx`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/*`, `frontend/src/pages/TranslationFlow.tsx`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `docs/frontend/translation-flow-persona-ui/spec.md`, `frontend/src/components/translation-flow/PersonaPanel.tsx`, `frontend/src/hooks/features/translationFlow/types.ts`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/components/translation-flow/PersonaPanel.tsx`]
- acceptance: [`summary/list/detail/footer が `empty` / `ready` / `cachedOnly` / `running` / `completed` / `partialFailed` / `failed` と整合する`, `cached/generated は persona 本文、pending/failed は未生成理由と会話抜粋を表示する`, `初回表示で backend 正常状態を `failed` 扱いしない`]
- [ ] 9.1 実装
- [ ] 9.2 検証

## 10. Translation Flow Page Wiring

- section_id: frontend-translationflow-page-wiring
- owner: frontend
- goal: `TranslationFlow` ページで revision 後の persona hook/panel 契約を配線し、load -> terminology -> persona -> summary の導線と初回表示回帰を閉じる。
- depends_on: [`frontend-translationflow-persona-hook`, `frontend-persona-panel-rendering`]
- shared_contract: [`TranslationFlow` tab order contract, `PersonaPanelProps`, `useTranslationFlow` actions contract]
- owned_paths: [`frontend/src/pages/TranslationFlow.tsx`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/*`, `frontend/src/components/translation-flow/*`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `docs/frontend/translation-flow-persona-ui/spec.md`, `frontend/src/pages/TranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/components/translation-flow/PersonaPanel.tsx`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/pages/TranslationFlow.tsx`, `cd frontend && npm run e2e -- src/e2e/translation-flow-required-scenarios.spec.ts`]
- acceptance: [`page が persona panel に revision 後の state/actions を渡す`, `タブ遷移が load -> terminology -> persona -> summary の順を保つ`, `persona 初回表示が `failed` へ倒れない`, `他 phase の panel wiring を壊さない`]
- [ ] 10.1 実装
- [ ] 10.2 検証
