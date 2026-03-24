# Bugfix Packet Template

正本は `changes/<id>/context_board/fix-distill.packet.json` とし、validation は同じ場所の
`fix-distill.packet.validation.json` を使う。

```json
{
  "invoked_skill": "fix-distill",
  "invoked_by": "fix-direction",
  "change": "<change-id>",
  "symptom": "",
  "repro_steps": [],
  "scope_summary": "",
  "observed_facts": [],
  "related_specs": [],
  "related_code": [],
  "active_constraints": [],
  "known_facts": [],
  "missing_observations": [],
  "unknowns": [],
  "current_scope": [],
  "next_action": "fix-trace",
  "state_summary_seed": {
    "reproduction_status": "",
    "known_facts": [],
    "unknowns": [],
    "active_logs": [],
    "current_scope": [],
    "next_action": "fix-trace"
  },
  "handoff": "fix-trace"
}
```

## Rules

- `observed_facts` には確認済みの事実だけを書く。
- `scope_summary` には現時点でどこまで絞れているかを書く。
- `missing_observations` には次の再現や tracing で見るべき点だけを書く。
- `state_summary_seed` は `fix-direction` がそのまま `State Summary` へ転記できる粒度で書く。
- `known_facts` `unknowns` `current_scope` `next_action` は validator の最低必須 field として毎回出力する。
- `handoff` には原則 `fix-trace` または `fix-direction` を書く。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
