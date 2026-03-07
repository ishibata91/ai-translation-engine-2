## ADDED Requirements

### Requirement: Task は進捗値を決定して Progress へ報告しなければならない
`task` は段階進捗の `phase/current/total/message` を決定し、`progress` へ報告しなければならない。`total` は全 request 件数を分母として扱わなければならない。

#### Scenario: フェーズ遷移時に task が進捗を報告する
- **WHEN** Task が enqueue/dispatch/save/complete の各段階へ遷移する
- **THEN** `task` は対応する `phase` と進捗値を `progress` へ報告しなければならない

#### Scenario: Progress は task が渡した値をそのまま配信する
- **WHEN** `task` が `phase/current/total/message` を `progress` へ報告する
- **THEN** UI へ配信される進捗値は task が指定した値と一致しなければならない

### Requirement: Task 状態遷移は task イベントで UI へ通知しなければならない
`task` の状態遷移（`pending/running/completed/failed/canceled`）は `task` イベント（例: `task:updated`）として UI に通知しなければならない。状態遷移通知は `progress` 経路を介してはならない。

#### Scenario: task 状態が UI へ直接通知される
- **WHEN** task 状態が `running` から `completed` に変化する
- **THEN** UI は `task` イベントで状態変更を受け取り、ステータス表示を更新しなければならない

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
`task` は MasterPersona のキュー投入後に再開コンテキスト（task_id, phase, resume_cursor）を保存し、再起動後およびキャンセル後の再開要求を受理しなければならない。再開時のモデル・パラメータは常に `config` から再読込しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に ResumeTask が呼ばれる
- **THEN** `task` は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** `task` は完了済み request を再実行せず、未完了 request のみを再開しなければならない
