# Current Context

## Bug
- change id: translation-flow-terminology-ui-progress
- symptom: データロード後に単語翻訳画面へ移動すると、対象単語リストが読込中のまま維持され、単語翻訳を実行できない
- repro request: 翻訳フローのデータロード完了後に単語翻訳 phase へ遷移したときの stuck 条件を特定し修正する

## Related Docs
- docs: docs/slice/terminology/terminology_test_spec.md
- changes: changes/translation-flow-terminology-ui-progress/ui.md, scenarios.md, logic.md, tasks.md

## Related Code
- files: pkg/workflow/translation_flow_service.go, pkg/workflow/translation_flow_service_test.go, frontend/src/hooks/features/translationFlow/useTranslationFlow.tsx, frontend/src/components/translation-flow/TerminologyPanel.tsx
- symbols: TranslationFlowService.LoadFiles, TranslationFlowService.ListTerminologyTargets, TranslationFlowService.GetTerminologyPhase

## Current Focus
- current hypothesis: LoadFiles 完了後も terminology phase summary の running 状態が残り、frontend が isRunning 扱いを継続していた
- next observation: 実アプリでデータロード直後の単語翻訳画面が pending/hidden に戻ることを確認する
