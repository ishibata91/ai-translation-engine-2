## ADDED Requirements

### Requirement: task スコープの shared handoff 成果物は task 削除時に cleanup されなければならない
システムは、`task_id` を親キーに保持する shared handoff 成果物について、task 削除時に当該 task の成果物だけを cleanup できなければならない。cleanup は task 管理の削除導線から呼び出され、他 task の handoff 成果物へ波及してはならない。

#### Scenario: translation flow handoff を task 単位で削除する
- **WHEN** task 管理が translation project task の削除中に handoff cleanup を要求する
- **THEN** システムは当該 `task_id` の file 親行と配下の section データを削除しなければならない
- **AND** 外部キーまたは同等の整合性制御で関連行の取り残しを発生させてはならない

#### Scenario: 別 task の handoff は保持される
- **WHEN** ある translation project task の handoff 成果物を cleanup する
- **THEN** システムは別 `task_id` の handoff 成果物を保持しなければならない
- **AND** 他 task の再表示や再開に必要なデータを削除してはならない
