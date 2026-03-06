## ADDED Requirements

### Requirement: Queue はタスク種別非依存で request 永続状態を保持しなければならない
Queue は特定スライスの知識を持たず、各 request について `pending/running/completed/failed/canceled` 状態と `resume_cursor` を既存テーブル拡張で永続化しなければならない。対象タスクの識別は `task_id` と `task_type` で行わなければならない。

#### Scenario: リクエスト保存後に再起動しても状態が失われない
- **WHEN** request を enqueue した後にアプリが再起動する
- **THEN** Queue は永続化された request 状態を復元し、再開対象を判定できなければならない

#### Scenario: 再開時に completed をスキップする
- **WHEN** QueueManager が任意タスクの再開を実行する
- **THEN** Queue は `completed` request を再送せず、`pending/failed/canceled` のみ再実行しなければならない
