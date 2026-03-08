## MODIFIED Requirements

### Requirement: Completed タスクに紐づく llm_queue job は削除されなければならない
Queue は、MasterPersona タスクが `Completed` に確定した後、workflow からの明示的な cleanup 呼び出しに応じて、同一 `task_id` に紐づく `llm_queue` job を全件削除しなければならない。

#### Scenario: Completed 遷移後に関連 job が全削除される
- **WHEN** workflow が MasterPersona タスク状態を `Completed` に確定し、queue cleanup を要求する
- **THEN** Queue は同一 `task_id` に紐づく `llm_queue` job をすべて削除しなければならない

#### Scenario: Failed または Canceled では job を保持する
- **WHEN** workflow が MasterPersona タスクを `Failed` または `Canceled` として終了する
- **THEN** Queue は再開や調査に必要な `llm_queue` job を削除してはならない
