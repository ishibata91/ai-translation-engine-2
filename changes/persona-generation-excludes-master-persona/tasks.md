## 1. Persona Config Contracts

- section_id: frontend-translationflow-persona-config-contracts
- owner: frontend
- status: completed
- goal: `useTranslationFlow` と `TranslationFlow` が terminology 設定と独立した persona 用 LLM/prompt state と action を公開できるよう、translation flow の公開型契約を固定する。
- depends_on: []
- shared_contract: [`UseTranslationFlowResult.state.personaConfig`, `UseTranslationFlowResult.state.personaPromptConfig`, `UseTranslationFlowResult.state.isPersonaConfigHydrated`, `UseTranslationFlowResult.state.isPersonaPromptHydrated`, `UseTranslationFlowResult.actions.handlePersonaConfigChange`, `UseTranslationFlowResult.actions.handlePersonaPromptChange`]
- condensed_brief:
  - why_now: hook と panel/page が persona 用 namespace と編集 UI を前提に進めるため、先に公開 contract を固定する必要がある。
  - fixed_contracts: persona phase は `translation_flow.persona.llm` / `translation_flow.persona.llm.<provider>` / `translation_flow.persona.prompt` を使い、terminology phase の state/action 名は維持する。
  - non_goals: backend DTO 変更、Wails binding 変更、panel 実装、hook の hydrate/persist ロジック実装。
  - known_blockers: なし。
  - validation_baseline: `frontend/src/hooks/features/translationFlow/types.ts` は terminology 契約を保持したまま persona 契約を追加できる。
  - carry_over_notes: 旧 `tasks.md` の section 7 は persona state machine 型までで止まっており、config state/action 契約は未定義として再計画する。
- owned_paths: [`frontend/src/hooks/features/translationFlow/types.ts`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`, `frontend/src/components/translation-flow/*`, `frontend/src/pages/TranslationFlow.tsx`, `frontend/src/e2e/*`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `changes/persona-generation-excludes-master-persona/logic.md`, `frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/masterPersona/useMasterPersona.tsx`, `frontend/src/components/ModelSettings.tsx`, `frontend/src/components/masterPersona/PromptSettingCard.tsx`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/hooks/features/translationFlow/types.ts`]
- acceptance: [`translation flow の公開 state/actions に persona 用 config/prompt と hydration flag が追加される`, `terminology 用 state/actions 名と型が壊れない`, `後続 section が `types.ts` だけを読めば persona 設定 contract を参照できる`]
- [x] 1.1 実装
- [x] 1.2 検証

## 2. Persona Hook Config Runtime

- section_id: frontend-translationflow-persona-hook-config-runtime
- owner: frontend
- status: completed
- goal: `useTranslationFlow` が persona 用 namespace を hydrate/persist し、初回は terminology 値から persona namespace へ移行したうえで、persona 実行時に persona 専用 request/prompt を使うようにする。
- depends_on: [`frontend-translationflow-persona-config-contracts`]
- shared_contract: [`translation_flow.persona.llm`, `translation_flow.persona.llm.<provider>`, `translation_flow.persona.prompt`, `persona namespace 未保存時は terminology 値を初回コピーしてから persona state を hydrate する`, `RunTranslationFlowPersona(taskID, request, prompt)` には persona config/prompt を渡す]
- condensed_brief:
  - why_now: change の主目的は persona phase の model/prompt 分離であり、hook が terminology 設定を流用している現状をここで止める必要がある。
  - fixed_contracts: backend の persona 実行 DTO は変更しない。provider 切替と debounce save は terminology と同等の運用を保つ。既存 task では persona namespace が空でも terminology 値から復元される。
  - non_goals: panel 見た目調整、page props 配線、backend/controller 実装。
  - known_blockers: `ConfigController` は `translation_flow.persona.*` の既定値を持たないため、hook 側で migration fallback を完結させる前提。
  - validation_baseline: 既存 `useTranslationFlow.test.tsx` には persona phase probe と terminology config namespace test があり、namespace 分離と run payload の退行を同ファイルで固定できる。
  - carry_over_notes: old plan の hook section は backend 依存を含んでいたが、今回 packet では frontend hook 内の保存戦略変更だけで閉じる。
- owned_paths: [`frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/translationFlow/adapters.ts`, `frontend/src/components/translation-flow/*`, `frontend/src/pages/TranslationFlow.tsx`, `frontend/src/e2e/*`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `changes/persona-generation-excludes-master-persona/logic.md`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`, `frontend/src/hooks/features/translationFlow/types.ts`, `frontend/src/hooks/features/masterPersona/useMasterPersona.tsx`, `frontend/src/hooks/features/masterPersona/helpers.ts`, `pkg/controller/config_controller.go`, `pkg/workflow/config/master_persona_prompt_defaults.go`, `docs/gateway/config/spec.md`]
- validation_commands: [`cd frontend && npm run test -- src/hooks/features/translationFlow/useTranslationFlow.test.tsx`, `cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/hooks/features/translationFlow/useTranslationFlow.tsx src/hooks/features/translationFlow/useTranslationFlow.test.tsx`]
- acceptance: [`persona phase 実行時に terminology ではなく persona namespace 由来の request/prompt が使われる`, `persona namespace 未保存の既存 task でも初回 hydrate で terminology 値が persona state に移行される`, `再表示時に persona config/prompt が復元される`, `terminology phase の保存挙動とテストが壊れない`]
- [x] 2.1 実装
- [x] 2.2 検証

## 3. Persona Panel Settings UI

- section_id: frontend-persona-panel-settings-ui
- owner: frontend
- status: completed
- goal: `PersonaPanel` に persona 用 model 設定と prompt 設定の編集 UI を追加し、state machine ごとの操作可否を維持したまま summary/list/detail と同居させる。
- depends_on: [`frontend-translationflow-persona-config-contracts`]
- shared_contract: [`PersonaPanel` は persona 用 `llmConfig` `promptConfig` `isConfigHydrated` `isPromptHydrated` `onConfigChange` `onPromptChange` を受け取る, settings UI は `ModelSettings` と `PromptSettingCard` を流用する, props 未配線時でも compile を壊さない default を持てる]
- condensed_brief:
  - why_now: UI contract で要求される settings card を panel 内に実装しないと、hook で分離した persona 設定を編集できない。
  - fixed_contracts: 既存 summary/list/detail/footer の state semantics は維持する。cached/generated は persona 本文表示、pending/failed は会話抜粋表示のままにする。
  - non_goals: hook の永続化ロジック、page からの props 配線、e2e mock 更新。
  - known_blockers: page 側配線前でもこの section 単体で lint/typecheck 可能なよう props は安全な default を持たせる必要がある。
  - validation_baseline: `PersonaPanel.tsx` は現状 settings placeholder 文のみで、独立 section として UI 差し替えが完結する。
  - carry_over_notes: old plan の panel section は rendering 回帰中心だったが、本 reroute では settings card 導入が主目的になる。
- owned_paths: [`frontend/src/components/translation-flow/PersonaPanel.tsx`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/*`, `frontend/src/pages/TranslationFlow.tsx`, `frontend/src/e2e/*`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `frontend/src/components/translation-flow/PersonaPanel.tsx`, `frontend/src/components/ModelSettings.tsx`, `frontend/src/components/masterPersona/PromptSettingCard.tsx`, `frontend/src/hooks/features/translationFlow/types.ts`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/components/translation-flow/PersonaPanel.tsx`]
- acceptance: [`persona panel に model 設定と prompt 設定 UI が追加される`, `settings UI が `ready` / `cachedOnly` / `completed` / `partialFailed` で表示され、running 中は編集/実行の可否が state に従う`, `summary/list/detail/footer の既存 state 表示を壊さない`]
- [x] 3.1 実装
- [x] 3.2 検証

## 4. Translation Flow Persona Page Wiring

- section_id: frontend-translationflow-persona-page-wiring
- owner: frontend
- status: completed
- goal: `TranslationFlow` ページと translation-flow E2E mock を persona 用 config contract に追従させ、panel への props 配線と required scenario の回帰確認基盤を揃える。
- depends_on: [`frontend-translationflow-persona-hook-config-runtime`, `frontend-persona-panel-settings-ui`]
- shared_contract: [`TranslationFlow` は persona panel に persona config/prompt/hydration/actions を渡す, e2e mock config store は `translation_flow.persona.llm` / `translation_flow.persona.llm.<provider>` / `translation_flow.persona.prompt` を返せる, required scenarios は translation flow 既存導線を壊さない]
- condensed_brief:
  - why_now: hook と panel をつないで mock 側も persona namespace を返せるようにしないと、実画面と E2E の両方で regression を検知できない。
  - fixed_contracts: tab 順は `load -> terminology -> persona -> summary` のまま。existing terminology mock 値は維持し、persona 用値は別 namespace で追加する。
  - non_goals: page object 拡張、backend mock/controller 実装追加、design docs 更新。
  - known_blockers: existing required scenario spec は persona assertions をまだ持たないため、この section では mock と page 配線の整合を優先し既存 suite の green を守る。
  - validation_baseline: `TranslationFlow.tsx` は props 配線のみ、`frontend/src/e2e/helpers/wails-mock.ts` と `frontend/src/e2e/fixtures/translation-flow/mock-data.ts` は translation flow 固有 mock を閉じ込めている。
  - carry_over_notes: old plan の page section から backend 依存を除去し、frontend mock/wiring に閉じる。
- owned_paths: [`frontend/src/pages/TranslationFlow.tsx`, `frontend/src/e2e/helpers/wails-mock.ts`, `frontend/src/e2e/fixtures/translation-flow/mock-data.ts`]
- forbidden_paths: [`frontend/src/hooks/features/translationFlow/*`, `frontend/src/components/translation-flow/*`, `frontend/src/e2e/page-objects/*`, `pkg/*`, `docs/*`, `changes/*`]
- required_reading: [`changes/persona-generation-excludes-master-persona/ui.md`, `changes/persona-generation-excludes-master-persona/scenarios.md`, `frontend/src/pages/TranslationFlow.tsx`, `frontend/src/components/translation-flow/PersonaPanel.tsx`, `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`, `frontend/src/e2e/helpers/wails-mock.ts`, `frontend/src/e2e/fixtures/translation-flow/mock-data.ts`, `frontend/src/e2e/translation-flow-required-scenarios.spec.ts`]
- validation_commands: [`cd frontend && npm run typecheck`, `cd frontend && npm run lint:file -- src/pages/TranslationFlow.tsx src/e2e/helpers/wails-mock.ts src/e2e/fixtures/translation-flow/mock-data.ts`, `cd frontend && npm run e2e -- src/e2e/translation-flow-required-scenarios.spec.ts`]
- acceptance: [`TranslationFlow が persona panel に persona config/prompt と change handlers を配線する`, `translation flow E2E mock が persona namespace の config 値を返し、既存 terminology namespace と衝突しない`, `required scenario suite が green のまま translation flow の既存導線を維持する`]
- [x] 4.1 実装
- [x] 4.2 検証
