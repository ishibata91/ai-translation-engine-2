# Bugfix Packet Template

```md
## Bugfix Packet
- symptom:
- repro_steps:
- scope_summary:
- observed_facts:
- related_specs:
- related_code:
- active_constraints:
- missing_observations:
- state_summary_seed:
  - reproduction_status:
  - known_facts:
  - unknowns:
  - active_logs:
  - current_scope:
  - next_action:
- handoff:
```

## Rules

- `observed_facts` には確認済みの事実だけを書く。
- `scope_summary` には現時点でどこまで絞れているかを書く。
- `missing_observations` には次の再現や tracing で見るべき点だけを書く。
- `state_summary_seed` は `fix-direction` がそのまま `State Summary` へ転記できる粒度で書く。
- `handoff` には原則 `fix-trace` または `fix-direction` を書く。
