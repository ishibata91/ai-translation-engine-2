## ADDED Requirements

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
`task` は MasterPersona のキュー投入後に再開コンテキスト（task_id, phase, resume_cursor）を保存し、再起動後およびキャンセル後の再開要求を受理しなければならない。再開時のモデル・パラメータは常に `config` から再読込しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に ResumeTask が呼ばれる
- **THEN** `task` は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** `task` は完了済み request を再実行せず、未完了 request のみを再開しなければならない
