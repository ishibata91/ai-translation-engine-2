## MODIFIED Requirements

### Requirement: Queue はタスク種別非依存で request 永続状態を保持しなければならない
Queue は特定スライスの知識を持たず、各 request について `pending/running/completed/failed/canceled` 状態と `resume_cursor` を既存テーブル拡張で永続化しなければならない。対象タスクの識別は `task_id` と `task_type` で行わなければならない。queue は workflow から契約経由で利用される runtime として振る舞い、controller、workflow、usecase slice の具体実装を知ってはならない。

#### Scenario: リクエスト保存後に再起動しても状態が失われない
- **WHEN** workflow が request を enqueue した後にアプリが再起動する
- **THEN** Queue は永続化された request 状態を復元し、再開対象を判定できなければならない

#### Scenario: 再開時に completed をスキップする
- **WHEN** workflow が任意タスクの再開を実行する
- **THEN** Queue は `completed` request を再送せず、`pending/failed/canceled` のみ再実行しなければならない

#### Scenario: Queue はスライス固有ロジックを持たない
- **WHEN** Queue が MasterPersona の request を保持または再開する
- **THEN** Queue は persona 固有の保存判定や UI 状態遷移を実装してはならない
- **AND** request 永続化と実行制御だけに集中しなければならない

### Requirement: Completed タスクに紐づく llm_queue job は削除されなければならない
Queue は、MasterPersona タスクが `Completed` に確定した後、workflow からの明示的な cleanup 呼び出しに応じて、同一 `task_id` に紐づく `llm_queue` job を全件削除しなければならない。完了済みタスクに対して再開不要な job を残してはならない。cleanup 判定は workflow が担い、queue が task state を自律的に判断してはならない。

#### Scenario: Completed 遷移後に関連 job が全削除される
- **WHEN** workflow が MasterPersona タスク状態を `Completed` に確定し、queue cleanup を要求する
- **THEN** Queue は同一 `task_id` に紐づく `llm_queue` job をすべて削除しなければならない
- **AND** 他タスクの job を削除してはならない

#### Scenario: Failed または Canceled では job を保持する
- **WHEN** workflow が MasterPersona タスクを `Failed` または `Canceled` として終了する
- **THEN** Queue は再開や調査に必要な `llm_queue` job を削除してはならない

#### Scenario: queue 削除後も完了結果の参照は継続できる
- **WHEN** `Completed` 後に関連 `llm_queue` job が削除された
- **THEN** システムはペルソナ本体や詳細表示に必要な永続データを保持し続けなければならない
