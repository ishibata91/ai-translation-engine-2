# Planning Packet Template

`plan-distill` の最終返答は Markdown で返す。
設計判断を始めるための packet に限定し、実装判断や worker 分割は混ぜない。

## Template

```md
## Planning Packet
- request:
- current_artifacts:
- relevant_specs:
- constraints:
- open_decisions:
- conflicts:
- draft_targets:
- handoff:
```

## Rules

- `current_artifacts` には `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無と鮮度を書く。
- `relevant_specs` には次の skill が読むべき docs とコードだけを書く。
- `constraints` には設計判断に影響する固定条件だけを書く。
- `open_decisions` には次の plan skill で決める論点だけを書く。
- `conflicts` には artifact 間の矛盾だけを書く。
- `handoff` には `next_skill` と `must_read` を含める。
- 長文引用は禁止する。
