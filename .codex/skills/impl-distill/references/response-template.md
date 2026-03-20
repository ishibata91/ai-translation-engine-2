# Implementation Packet Template

```md
## Implementation Packet
- task:
- scope:
- must_read:
- interfaces:
- entry_points:
- edit_boundary:
- owned_paths_candidates:
- quality_gates:
- constraints:
- acceptance:
- unknowns:
- handoff:
```

## Rules

- `interfaces` には worker が守る contract だけを書く。
- `entry_points` には path と symbol を短く書く。
- `edit_boundary` には変更可能範囲と触ってはいけない範囲を分ける。
- `unknowns` は実装判断を止めるものだけに限定する。
- `handoff` には原則 `impl-workplan` を書く。
