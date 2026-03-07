## ADDED Requirements

### Requirement: Completed タスクに紐づく llm_queue job は削除されなければならない
Queue は、MasterPersona タスクが `Completed` に確定した後、同一 `task_id` に紐づく `llm_queue` job を全件削除しなければならない。完了済みタスクに対して再開不要な job を残してはならない。

#### Scenario: Completed 遷移後に関連 job が全削除される
- **WHEN** MasterPersona タスク状態が `Completed` に遷移する
- **THEN** システムは同一 `task_id` に紐づく `llm_queue` job をすべて削除しなければならない
- **AND** 他タスクの job を削除してはならない

#### Scenario: Failed または Canceled では job を保持する
- **WHEN** MasterPersona タスクが `Failed` または `Canceled` で終了する
- **THEN** システムは再開や調査に必要な `llm_queue` job を削除してはならない

#### Scenario: queue 削除後も完了結果の参照は継続できる
- **WHEN** Completed 後に関連 `llm_queue` job が削除された
- **THEN** システムはペルソナ本体や詳細表示に必要な永続データを保持し続けなければならない
