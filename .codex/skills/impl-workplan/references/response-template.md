# Workplan Packet Template

```md
## Workplan Packet
- change:
- tasks_path:
- progress_snapshot:
- shared_contracts:
- dispatch_order:
- sections:
- work_orders:
- unresolved:
- handoff:
```

## Rules

- `sections` には 1 section = 1 owner の plan だけを書く。
- `sections` は `references/templates.md` の `Section Plan` schema を省略なく使う。
- `sections` と `work_orders` の validation field は `validation_commands` のみとし、他名義の validation field を混在させない。
- `work_orders` は `references/templates.md` の `Work Order` schema をそのまま使い、`impl-frontend-work` または `impl-backend-work` が追加判断なしで受け取れる形にする。
- `progress_snapshot` には各 section の最新 status を含める。
- `tasks.md` の section 契約と初期 status は `impl-workplan` が定義し、runtime 中の status / 実装 / 検証更新だけを `impl-direction` が行う。
- `unresolved` は worker 起動を止めるものだけに限定する。
- `unresolved` には owner 未確定、shared contract 未固定、required field 欠落を含める。
- `handoff` には原則 `impl-direction` を書く。
