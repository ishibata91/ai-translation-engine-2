# Design Review

score: 0.91

### Design Review Findings
- なし

### Open Questions
- `fix` lane の state summary を skill 応答 packet のまま扱うか、固定ファイルへ昇格するかは実装時に決める必要がある
- `impl-workplan` の `tasks.md format` へ `completed_with_noise` をどう表現するかは template 更新時に詰める必要がある

### Residual Risks
- `fix-review` は 7 field schema を維持するため、未解消 scope と外部ノイズの分離が `required_delta` / `recheck` の構造化品質に依存する
- `impl-direction` が `tasks.md` を正本化しても、既存 change の古い `tasks.md` との後方互換方針を決めないと移行時に揺れやすい

### Docs Sync
- 要否: 不要。まずは `changes/impl-fix-skill-ops-hardening/` で差分仕様として保持し、skill 実装完了後に `.codex/skills/` の正本更新可否を判断する
