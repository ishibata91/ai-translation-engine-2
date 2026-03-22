# Impl Workplan Templates

`Section Plan` `Work Order` `tasks.md format` は相互に同じ section schema を共有する。
`tasks.md` の section 契約は `impl-workplan` が定義し、worker skill はこの format を前提に読むだけとする。

## Section Plan
```md
- section_id:
- title:
- owner: frontend | backend
- status: pending | in_progress | completed | blocked | completed_with_noise
- goal:
- depends_on:
- shared_contract:
- condensed_brief:
  why_now:
  fixed_contracts:
  non_goals:
  known_blockers:
  validation_baseline:
  carry_over_notes:
- owned_paths:
- forbidden_paths:
- required_reading:
- validation_commands:
- acceptance:
```

## Work Order
```md
### Work Order
- change:
- section_id:
- title:
- owner: frontend | backend
- status:
- goal:
- depends_on:
- progress_snapshot:
- shared_contract:
- condensed_brief:
  why_now:
  fixed_contracts:
  non_goals:
  known_blockers:
  validation_baseline:
  carry_over_notes:
- required_reading:
- owned_paths:
- forbidden_paths:
- validation_commands:
- acceptance:
```

## tasks.md format
```md
## 1. <Section Title>

- section_id: <stable-section-id>
- owner: frontend | backend
- status: pending | in_progress | completed | blocked | completed_with_noise
- goal:
- depends_on:
- shared_contract:
- condensed_brief:
  - why_now:
  - fixed_contracts:
  - non_goals:
  - known_blockers:
  - validation_baseline:
  - carry_over_notes:
- owned_paths:
- forbidden_paths:
- required_reading:
- validation_commands:
- acceptance:
- [ ] 1.1 実装
- [ ] 1.2 検証
```

## Rules

- `Section Plan` `Work Order` `tasks.md format` は `validation_commands` を唯一の validation field として共有する。
- `owner` は `frontend | backend` のどちらかで確定済みの値だけを書く。
- `shared_contract` は worker 起動前に固定済みであり、かつ worker が `owned_paths` 内だけで section を完了するために必要十分な契約だけを書く。
- `condensed_brief` は worker が履歴全文を再読せずに着手するための圧縮本文であり、`why_now` `fixed_contracts` `non_goals` `known_blockers` `validation_baseline` `carry_over_notes` を欠かさない。
- owner 未確定、shared contract 未固定、required field 欠落の項目は section にせず `unresolved` 側で止める。
- `tasks.md` の section 契約と初期 status は `impl-workplan` が定義し、実装中の status / 実装 / 検証注記だけを `impl-direction` が更新する。
- constructor / DI / test stub / compile dependency が `owned_paths` 外に必要な場合は、該当 section をそのまま生成せず先行 section を分離するか `unresolved` に倒す。
- `depends_on` に書かれた先行 section を実装しても downstream worker が blocked になるなら、section plan は未完成として扱う。
