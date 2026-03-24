---
title: Persona Phase Shared Npc List
description: ペルソナ生成 phase の NPC 一覧を MasterPersona 一覧と共有化する差分仕様
---

# Persona Phase Shared Npc List

## Documents
- [ui](./ui/)
- [scenarios](./scenarios/)
- [logic](./logic/)
- [review](./review/)

## Scope
- Translation Flow の `ペルソナ生成` phase にある NPC 一覧を MasterPersona 一覧と同じ表形式へ揃える
- 両画面が同じ frontend list component を共有できるように UI 契約と責務境界を定義する
- 詳細ペインや phase 実行制御は各画面の責務として維持する

## Next Steps
- `ui.md` で共有一覧コンポーネントの観測可能な UI 事実を確定する
- `scenarios.md` で translation flow 側の選択、再表示、再試行時の受け入れ条件を固定する
- `logic.md` をもとに `impl-direction` へ handoff する
