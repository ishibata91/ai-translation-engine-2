## MODIFIED Requirements

### Requirement: Queue はタスク種別非依存で request 永続状態を保持しなければならない
Queue は特定スライスの知識を持たず、各 request について `pending/running/completed/failed/canceled` 状態と `resume_cursor` を既存テーブル拡張で永続化しなければならない。対象タスクの識別は `task_id` と `task_type` で行わなければならない。queue は workflow から契約経由で利用される runtime として振る舞い、controller、workflow、usecase slice の具体実装を知ってはならない。

#### Scenario: リクエスト保存後に再起動しても状態が失われない
- **WHEN** workflow が request を enqueue した後にアプリが再起動する
- **THEN** Queue は永続化された request 状態を復元し、再開対象を判定できなければならない

#### Scenario: Queue は runtime として振る舞う
- **WHEN** Queue が request を dispatch する
- **THEN** Queue は実行制御だけを担わなければならない
- **AND** slice 固有の保存判定を持ってはならない

### Requirement: Queue worker は gateway を通じて外部依頼を行わなければならない
Queue worker は LLM などの外部依頼を runtime から直接実装してはならず、gateway 契約へ委譲しなければならない。

#### Scenario: queue worker は LLM gateway を利用する
- **WHEN** queue worker が request を外部 LLM へ送る
- **THEN** worker は gateway 契約を通じて依頼しなければならない
- **AND** worker 自身が slice 固有レスポンス解釈をしてはならない
