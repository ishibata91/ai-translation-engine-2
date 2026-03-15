## ADDED Requirements

### Requirement: Task 管理は task 削除時に task manager 管理対象だけを整理しなければならない
Task 管理は、task 削除要求を受けたとき、frontend task の永続レコード削除と task manager が保持する管理対象の整理を同一の削除導線で実行しなければならない。task 削除は artifact 正本や共有成果物を削除してはならない。

#### Scenario: 停止済み task を削除する
- **WHEN** ユーザーが `pending`、`paused`、`request_generated`、`failed`、`cancelled` のいずれかの status を持つ task の削除を要求する
- **THEN** システムは `frontend_tasks` の当該 task レコードを削除しなければならない
- **AND** task manager が保持する当該 task の管理対象を整理しなければならない
- **AND** artifact 正本を削除してはならない

#### Scenario: 実行中 task は削除できない
- **WHEN** ユーザーが `running` status の task の削除を要求する
- **THEN** システムは task 削除を拒否しなければならない
- **AND** ユーザーには停止後に削除できることを通知しなければならない
