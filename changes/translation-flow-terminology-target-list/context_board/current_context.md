# Current Context

## Change
- change id: `translation-flow-terminology-target-list`
- request: `TerminologyPanel` に対象単語リストを追加し、`データロード` phase では同名ファイルの重複読み込みをブロックする。

## Related Docs
- changes:
  - `changes/translation-flow-terminology-target-list/index.md`
  - `changes/translation-flow-terminology-target-list/logic.md`
  - `changes/translation-flow-terminology-target-list/scenarios.md`
  - `changes/translation-flow-terminology-target-list/ui.md`
- docs:
  - `docs/workflow/translation-flow-data-load/spec.md`
  - `docs/frontend/translation-flow-data-load-ui/spec.md`

## Related Code
- backend:
  - `pkg/workflow/translation_flow.go`
  - `pkg/workflow/translation_flow_service.go`
  - `pkg/controller/task_controller.go`
  - `pkg/artifact/translationinput/`
- frontend:
  - `frontend/src/components/translation-flow/LoadPanel.tsx`
  - `frontend/src/components/translation-flow/TerminologyPanel.tsx`
  - `frontend/src/components/DataTable.tsx`
  - `frontend/src/hooks/features/translationFlow/types.ts`
  - `frontend/src/hooks/features/translationFlow/adapters.ts`
  - `frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx`
  - `frontend/src/hooks/features/translationFlow/useTranslationFlow.test.tsx`
  - `frontend/src/e2e/page-objects/pages/translation-flow.po.ts`
  - `frontend/src/e2e/translation-flow-required-scenarios.spec.ts`
  - `frontend/src/e2e/fixtures/translation-flow/mock-data.ts`
- symbols:
  - `workflow.TranslationFlow`
  - `workflow.TerminologyPhaseResult`
  - `controller.TaskController`
  - `useTranslationFlow`
  - `TerminologyPanel`
  - `LoadPanel`

## Constraints
- fixed constraints:
  - `context_board/` を shared handoff の正本にする。
  - `context_manager` で context を圧縮してから coder に渡す。
  - `server-filesystem` を優先して検索・読取・書込する。
  - task は frontend / backend に分割し、単一 skill に混在させない。
- open constraints:
  - backend の preview API は `translationinput` artifact 正本から直接投影するのが最有力だが、公開面の置き場所を最終確定する必要がある。
  - frontend は preview state を summary と分離して持つ必要がある。

## Current Focus
- current design section: implementation orchestration の再初期化
- next question: frontend 側の preview state / duplicate block 実装と docs sync に進めるか
