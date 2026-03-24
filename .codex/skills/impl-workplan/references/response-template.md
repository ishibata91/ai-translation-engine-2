# Workplan Packet Template

正本は `changes/<id>/context_board/impl-workplan.packet.json` とし、validation は同じ場所の
`impl-workplan.packet.validation.json` を使う。

```json
{
  "change": "<change-id>",
  "tasks_path": "changes/<id>/tasks.md",
  "progress_snapshot": [],
  "shared_contracts": [],
  "dispatch_order": [],
  "sections": [],
  "work_orders": [],
  "unresolved": [],
  "handoff": "impl-direction"
}
```

## Rules

- `sections` には 1 section = 1 owner の plan だけを書く。
- `sections` は `references/templates.md` の `Section Plan` schema を省略なく使う。
- `sections` と `work_orders` の validation field は `validation_commands` のみとし、他名義の validation field を混在させない。
- `work_orders` は `references/templates.md` の `Work Order` schema をそのまま使い、`impl-frontend-work` または `impl-backend-work` が追加判断なしで受け取れる形にする。
- `progress_snapshot` には各 section の最新 status を含める。
- `tasks.md` の section 契約と初期 status は `impl-workplan` が定義し、runtime 中の status / 実装 / 検証更新だけを `impl-direction` が行う。
- `change` `tasks_path` `progress_snapshot` `shared_contracts` `dispatch_order` `sections` `unresolved` は validator の最低必須 field として毎回出力する。
- `unresolved` は worker 起動を止めるものだけに限定する。
- `unresolved` には owner 未確定、shared contract 未固定、required field 欠落を含める。
- `handoff` には原則 `impl-direction` を書く。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
