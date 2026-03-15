## ADDED Requirements

### Requirement: MasterPersona artifact は task 削除要求に対して task 中間生成物だけを cleanup できなければならない
`pkg/artifact/master_persona_artifact` は、task 完了時の cleanup に加えて、明示的な task 削除要求に対して `task_id` 単位の中間生成物 cleanup を提供しなければならない。cleanup は final 成果物を削除対象に含めてはならない。

#### Scenario: task 削除で temp 成果物を削除する
- **WHEN** task 管理が MasterPersona task の削除処理中に `task_id` cleanup を要求する
- **THEN** システムは当該 `task_id` の temp 成果物だけを削除しなければならない
- **AND** final persona 成果物は保持しなければならない

#### Scenario: 対象 task が存在しなくても cleanup が破綻しない
- **WHEN** task 管理が既に temp 成果物のない `task_id` に対して cleanup を要求する
- **THEN** システムはエラーなく cleanup を完了できなければならない
- **AND** 他 task の temp 成果物へ影響してはならない
