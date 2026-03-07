## MODIFIED Requirements

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
`task` は MasterPersona のキュー投入後に再開コンテキスト（`task_id`, `phase`, `resume_cursor`）を保存し、再起動後およびキャンセル後の再開要求を受理しなければならない。さらに開始時に受け取った `overwrite_existing` を task metadata に保持し、再開時も同一方針を必ず適用しなければならない。再開時のモデル・パラメータは常に `config` から再読込しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に ResumeTask が呼ばれる
- **THEN** `task` は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** `task` は完了済み request を再実行せず、未完了 request のみを再開しなければならない

#### Scenario: 再開時に上書き方針が保持される
- **WHEN** `overwrite_existing=false` で開始した MasterPersona タスクを再開する
- **THEN** `task` は開始時 metadata の `overwrite_existing` を再利用し、再開時に `true` として扱ってはならない

#### Scenario: REQUEST_GENERATED 後の自動再開要求を受理できる
- **WHEN** UI が `REQUEST_GENERATED` 直後に同一 task ID で `ResumeTask` を呼び出す
- **THEN** `task` はその再開要求を受理し、未完了 request の dispatch/save フェーズへ遷移しなければならない
