# Current Context

## Change
- change id: persona-generation-excludes-master-persona
- request: 翻訳フローの `ペルソナ生成` phase で、翻訳対象に登場する NPC の persona を生成できるようにしつつ、既存 Master Persona を持つ NPC は生成対象から除外する

## Related Docs
- changes:
  - `.codex/changes/persona-generation-excludes-master-persona/ui.md`
  - `.codex/changes/persona-generation-excludes-master-persona/scenarios.md`
  - `.codex/changes/persona-generation-excludes-master-persona/logic.md`
  - `.codex/changes/persona-generation-excludes-master-persona/review.md`
- docs:
  - `docs/frontend/translation-flow-persona-ui/spec.md`
  - `docs/workflow/translation-flow-persona-phase/spec.md`
  - `docs/slice/persona/spec.md`
  - `docs/slice/persona/npc_personaerator_test_spec.md`

## Related Code
- files:
  - `frontend/src/components/translation-flow/PersonaPanel.tsx`
  - `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`
  - `pkg/workflow/translation_flow.go`
  - `pkg/workflow/translation_flow_service.go`
  - `pkg/artifact/translationinput/contract.go`
  - `pkg/slice/persona/store.go`
- symbols:
  - `workflow.TranslationFlow`
  - `workflow.TranslationFlowService`
  - `translationinput.Repository`
  - `master_persona_artifact.LookupKey`

## Constraints
- fixed constraints:
  - 除外判定は `master_persona_artifact` final を正本とする
  - lookup key は `source_plugin + speaker_id` を使う
  - preview と execute は同じ candidate planner を使う
  - all-cached のときは no-op 完了とする
- open constraints:
  - `partialFailed` 後に後続 phase が persona 欠損行をどう扱うかは実装時に明示する必要がある
  - translation input artifact へ persona phase 用の candidate / summary API を追加する必要がある

## Current Focus
- current design section: 実装 handoff 用の設計差分と docs sync は完了
- next question: `impl-direction` で frontend / workflow / slice / artifact の実装順と owned paths を確定する
