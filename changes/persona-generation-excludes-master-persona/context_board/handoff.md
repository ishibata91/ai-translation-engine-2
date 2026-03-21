# Handoff

## Current Role
- role: design orchestrator
- skill: plan-direction

## Next Role
- role: implementation orchestrator
- skill: impl-direction

## Done
- confirmed decisions:
  - 翻訳フロー `ペルソナ生成` phase の UI 契約、主シナリオ、責務境界を change docs に確定した
  - `source_plugin + speaker_id` を使った既存 Master Persona 除外規則を docs 正本へ昇格した
  - all-cached no-op 完了、duplicate candidate 統合、failed 行だけの retry 規則を docs へ反映した

## Pending
- unresolved items:
  - translation flow 用の persona preview / summary / run API の具象契約をコードへ追加する
  - translation input artifact から persona candidate を取り出す保存境界を実装する
  - `partialFailed` 後に後続 phase へ進むときの translator fallback を実コードで担保する
- completion condition:
  - persona phase の preview / execute / retry が同じ candidate planner で動き、既存 Master Persona 行を再送しない実装とテストが揃うこと
