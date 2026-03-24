# Implementation Packet Template

正本は `changes/<id>/context_board/impl-distill.packet.json` とし、validation は同じ場所の
`impl-distill.packet.validation.json` を使う。

```json
{
  "invoked_skill": "impl-distill",
  "invoked_by": "impl-direction",
  "change": "<change-id>",
  "task": "",
  "scope": [],
  "must_read": [],
  "interfaces": [],
  "entry_points": [],
  "module_candidates": [],
  "shared_contract_candidates": [],
  "edit_boundary": {
    "allowed": [],
    "forbidden": []
  },
  "validation_commands": [],
  "constraints": [],
  "acceptance": [],
  "known_facts": [],
  "unknowns": [],
  "current_scope": [],
  "next_action": "impl-workplan",
  "handoff": "impl-workplan"
}
```

## Rules

- `interfaces` には worker が守る contract だけを書く。
- `entry_points` には path と symbol を短く書く。
- `module_candidates` には `impl-workplan` が section へ切り出せる単位を書く。
- `shared_contract_candidates` には section 着手前に固定すべき型・API 契約だけを書く。
- `edit_boundary` には変更可能範囲と触ってはいけない範囲を分ける。
- validation field は `validation_commands` だけを使い、`quality_gates` や他の validation 名は出力しない。
- `known_facts` `unknowns` `current_scope` `next_action` は validator の最低必須 field として毎回出力する。
- `unknowns` は実装判断を止めるものだけに限定する。
- `handoff` には原則 `impl-workplan` を書き、`tasks.md` 生成責務を戻さない。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
