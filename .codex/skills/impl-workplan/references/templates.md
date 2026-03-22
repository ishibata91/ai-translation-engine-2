# Impl Workplan Templates

`Section Plan` `Work Order` `tasks.md format` は相互に同じ section schema を共有する。
`tasks.md` は `impl-workplan` だけが生成または更新し、worker skill はこの format を前提に読むだけとする。

## Section Plan
```md
- section_id:
- title:
- owner: frontend | backend
- goal:
- depends_on:
- shared_contract:
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
- goal:
- depends_on:
- shared_contract:
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
- goal:
- depends_on:
- shared_contract:
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
- owner 未確定、shared contract 未固定、required field 欠落の項目は section にせず `unresolved` 側で止める。
- `tasks.md` の生成または更新は `impl-workplan` だけが行い、worker skill へ委譲しない。
- constructor / DI / test stub / compile dependency が `owned_paths` 外に必要な場合は、該当 section をそのまま生成せず先行 section を分離するか `unresolved` に倒す。
- `depends_on` に書かれた先行 section を実装しても downstream worker が blocked になるなら、section plan は未完成として扱う。
