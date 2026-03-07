## ADDED Requirements

### Requirement: Progress は task から受け取った進捗を中継しなければならない
`progress` は特定スライスの知識を持たず、task/オーケストレーション層から受け取った `task_id`、`task_type`、`phase`、`current`、`total`、`message` を中継しなければならない。`phase` の語彙定義と進捗値の決定は呼び出し側が管理しなければならない。

#### Scenario: 任意タスクの phase 進捗をそのまま通知できる
- **WHEN** 呼び出し側が任意の `phase` を指定して進捗を報告する
- **THEN** `progress` は phase 名を解釈せず、受け取った値をそのまま中継しなければならない

#### Scenario: 分母は呼び出し側が与えた total を使用する
- **WHEN** 呼び出し側が `current` と `total` を指定して進捗を報告する
- **THEN** `progress` は `total` を内部で再計算せず、指定値を配信しなければならない

#### Scenario: タスク状態遷移は progress で扱わない
- **WHEN** task の `pending/running/completed/failed/canceled` が更新される
- **THEN** `progress` は状態遷移イベントを生成または中継してはならない
