## MODIFIED Requirements

### Requirement: MasterPersona開始操作はペルソナリクエスト生成タスクを起動しなければならない
`MasterPersona` 画面の開始ボタンは、`task.Bridge` 経由で「マスターペルソナ生成タスク」を起動しなければならない。開始時入力には `source_json_path` と `overwrite_existing` を含めなければならない。システムはタスクIDを発行し、同一IDで進捗・完了・失敗を追跡可能でなければならない。

#### Scenario: 開始ボタン押下で上書き指定付きタスクが起動される
- **WHEN** ユーザーが `MasterPersona` 画面で開始ボタンを押下し、重複時上書きチェックを指定する
- **THEN** システムは `task.Bridge` に `source_json_path` と `overwrite_existing` を含む開始要求を送らなければならない
- **AND** タスクIDを発行して実行状態を `Running` に更新しなければならない

#### Scenario: タスクが完了しUI状態が更新される
- **WHEN** `PreparePrompts` が正常終了する
- **THEN** システムはタスクステータスを `REQUEST_GENERATED` に更新しなければならない
- **AND** UIは生成中表示を解除し、進捗表示と状態メッセージを更新可能な状態にしなければならない

#### Scenario: リクエスト生成完了後に自動でキュー実行へ遷移する
- **WHEN** `task:phase_completed` で `REQUEST_GENERATED` が通知される
- **THEN** UI は同一 task ID に対して `ResumeTask` を自動呼び出ししなければならない
- **AND** ユーザーの手動再開操作を必須としてはならない
