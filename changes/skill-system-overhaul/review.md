# Design Review

score: 0.92

### Design Review Findings
- [high] `aite2-ui-polish-orchestrate` が user-facing orchestration skill として書かれており、AGENTS.md の「入口は 4 direction skill のみ」と衝突している。加えて `plan-direction` の routing 例では既存見た目修正を `impl-direction` へ handoff しており、入口設計が二重化している。
- [high] impl lane の正本 workflow が二重化している。`impl-direction` は `implementer` 直実装を正本としている一方、`impl-distill` と `impl-frontend-work` / `impl-backend-work` は `impl-workplan` 前提で書かれている。新しい change plan では `impl-workplan` と専用 `workplan_builder` agent を新設し、`tasks.md` 生成と section planning を impl lane に固定する必要がある。
- [high] fix logging lifecycle が閉じていない。`fix-direction` は最後に `fix-logging` で一時ログ削除を要求するが、`fix-logging` は追加専用で削除契約を持たず、prefix も skill と agent で不一致になっている。
- [medium] `fix-direction` は `score < 0.85` で loop を制御するが、`fix-review` template には `score` field がない。review schema と orchestration 条件がずれている。
- [medium] investigation skill 群に `find_by_name` `view_file` `grep` など現行運用ルールにない tool 名が残っており、tool 規約が統一されていない。
- [medium] `impl-direction` の `ui.md がない時はフロントエンド実装なし` ルールは、frontend 影響を持つ change を backend-only と誤分類しうる。`impl-workplan` 導入後も artifact readiness の判定軸は routing matrix ベースへ改める必要がある。

### Open Questions
- UI polish 用 specialized flow を `impl-direction` 配下に統合したうえで helper として残すか、それとも完全に廃止するか

### Residual Risks
- hotfix だけ先に当てて全体 schema を揃えないと、別 lane に同種の不整合が残る
- `impl-workplan` の section schema を曖昧にしたまま導入すると、`workplan_builder` と worker の間で責務漏れが再発する
- deprecated skill を残置すると、将来また誤って user-facing に使われる可能性がある

### Docs Sync
- 要否: 不要。今回の成果物は `changes/skill-system-overhaul/` に閉じた運用設計差分として保持する
