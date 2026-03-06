## ADDED Requirements

### Requirement: Progress はタスク種別非依存の段階進捗を通知しなければならない
`progress` は特定スライスの知識を持たず、`task_id`、`task_type`、`phase`、`current`、`total`、`message` を受け取って配信しなければならない。`phase` の語彙定義は呼び出し側（task/オーケストレーション層）が管理しなければならない。

#### Scenario: 任意タスクの phase 進捗をそのまま通知できる
- **WHEN** 呼び出し側が任意の `phase` を指定して進捗を報告する
- **THEN** `progress` は phase 名を解釈せず、受け取った値をそのまま通知しなければならない

#### Scenario: 分母は呼び出し側が与えた total を使用する
- **WHEN** 呼び出し側が `current` と `total` を指定して進捗を報告する
- **THEN** `progress` は `total` を内部で再計算せず、指定値を配信しなければならない
