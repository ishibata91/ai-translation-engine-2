# Handoff

## Current Role
- role: context_manager
- skill: `default`

## Next Role
- role: implementation driver
- skill: `aite2-implementation-driver`

## Done
- confirmed decisions:
  - この change は mixed task として扱う。
  - shared board を追加し、以後の handoff は board 正本で進める。
  - frontend では `TerminologyPanel` の対象単語リスト追加とデータロード重複ブロックが主作業になる。
  - backend では terminology preview DTO の追加と summary/API 境界整理が主作業になる。
  - backend 実装は完了し、`backend:lint:file` と `go test ./pkg/...` を通過した。

## Pending
- unresolved items:
  - backend coder の所有範囲: `translation_flow.go` / `translation_flow_service.go` / `task_controller.go` に加えて `translationinput` artifact の公開面が必要か
  - frontend coder の所有範囲: `LoadPanel.tsx` / `TerminologyPanel.tsx` / hook / adapter / E2E のどこまでを一括で持たせるか
  - docs 同期を implementation 後に `aite2-sync-docs` へ回すかどうか
- completion condition:
  - backend と frontend の task 範囲が分離され、coder 起動に必要な入力が board 上で明示されること
