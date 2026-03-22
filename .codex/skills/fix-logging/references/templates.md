# Fix Logging Templates

## add packet
```md
### Fix Logging Operation
- operation: add
- log_prefix: `[fix-trace]`
- target_files:
- observation_points:
- state_summary_ref:
- reproduce_steps:
- expected_observation:
```

## remove packet
```md
### Fix Logging Operation
- operation: remove
- log_prefix: `[fix-trace]`
- target_files:
- log_additions:
- state_summary_ref:
- cleanup_reason: final accept reached
```

## result packet
```md
### Fix Logging Result
- operation:
- modified_files:
- log_additions:
- log_removals:
- state_summary_update:
- reproduce_steps:
- notes:
```

> `add` は `log_additions` を返し、`remove` は `log_removals` を返す。両 operation とも prefix は `[fix-trace]` に固定する。
