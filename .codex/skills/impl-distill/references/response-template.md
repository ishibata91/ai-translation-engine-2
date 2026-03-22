# Implementation Packet Template

```md
## Implementation Packet
- task:
- scope:
- must_read:
- interfaces:
- entry_points:
- module_candidates:
- shared_contract_candidates:
- edit_boundary:
- validation_commands:
- constraints:
- acceptance:
- unknowns:
- handoff:
```

## Rules

- `interfaces` には worker が守る contract だけを書く。
- `entry_points` には path と symbol を短く書く。
- `module_candidates` には `impl-workplan` が section へ切り出せる単位を書く。
- `shared_contract_candidates` には section 着手前に固定すべき型・API 契約だけを書く。
- `edit_boundary` には変更可能範囲と触ってはいけない範囲を分ける。
- validation field は `validation_commands` だけを使い、`quality_gates` や他の validation 名は出力しない。
- `unknowns` は実装判断を止めるものだけに限定する。
- `handoff` には原則 `impl-workplan` を書き、`tasks.md` 生成責務を戻さない。
