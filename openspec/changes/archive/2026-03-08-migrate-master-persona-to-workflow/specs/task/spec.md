## MODIFIED Requirements

### Requirement: Wailsバインディングの拡張
Task 管理機能は、controller 層が従来の `GetActiveTasks`, `CancelTask`, `ResumeTask` などの Wails バインディングを提供しつつ、MasterPersona の実際のユースケース進行は workflow 契約へ委譲しなければならない。

#### Scenario: Wails バインディングは workflow を経由して再開する
- **WHEN** UI から `ResumeTask` が呼び出される
- **THEN** controller は workflow の再開契約を呼び出さなければならない

#### Scenario: 既存の公開 API は維持される
- **WHEN** フロントエンドが既存の `GetActiveTasks`, `CancelTask`, `ResumeTask` を利用する
- **THEN** システムは同等の公開 API を継続提供しなければならない

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
MasterPersona タスクは workflow 主導で、`task_id`, `phase`, `resume_cursor`, `overwrite_existing` を保持し、再起動後およびキャンセル後の再開要求を受理しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に `ResumeTask` が呼ばれる
- **THEN** workflow は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** workflow は完了済み request を再実行せず、未完了 request のみを再開しなければならない
