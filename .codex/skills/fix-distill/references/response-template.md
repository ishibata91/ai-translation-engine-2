# Bugfix Packet Template

```md
## Bugfix Packet
- symptom:
- repro_steps:
- observed_facts:
- related_specs:
- related_code:
- missing_observations:
- handoff:
```

## Rules

- `observed_facts` には確認済みの事実だけを書く。
- `missing_observations` には次の再現や tracing で見るべき点だけを書く。
- `handoff` には原則 `fix-trace` または `fix-direction` を書く。
