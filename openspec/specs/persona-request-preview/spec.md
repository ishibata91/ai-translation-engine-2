# Persona Request Preview 仕様書

## 概要
`MasterPersona` 画面からペルソナ生成Phase 1（リクエスト生成）を起動し、タスク状態と既存ログビューアで結果確認できるようにする機能。

## ADDED Requirements

### Requirement: MasterPersona 画面の開始ボタンは task 境界を起動しなければならない
`MasterPersona` 画面の開始ボタンは、controller を通じて workflow 管理下の MasterPersona タスク開始要求を送らなければならない。開始時入力には `source_json_path` と `overwrite_existing` を含めなければならない。システムは task ID を発行し、同一 ID で進捗・完了・失敗を追跡可能でなければならない。

#### Scenario: UI が workflow 管理下の開始要求を送る
- **WHEN** ユーザーが MasterPersona 画面で開始ボタンを押す
- **THEN** システムは controller を通じて workflow 管理下の開始要求を送らなければならない
- **AND** 同一 task ID で進捗・完了・失敗を追跡できなければならない

#### Scenario: タスクが完了しUI状態が更新される
- **WHEN** `PreparePrompts` が正常終了する
- **THEN** システムはタスクステータスを `REQUEST_GENERATED` に更新しなければならない
- **AND** UIは生成中表示を解除し、進捗表示と状態メッセージを更新可能な状態にしなければならない

#### Scenario: リクエスト生成完了後に自動でキュー実行へ遷移する
- **WHEN** `task:phase_completed` で `REQUEST_GENERATED` が通知される
- **THEN** UI は同一 task ID に対して `ResumeTask` を自動呼び出ししなければならない
- **AND** ユーザーの手動再開操作を必須としてはならない

### Requirement: 生成結果の確認は既存log-viewerのinfoログで行わなければならない
ペルソナリクエスト生成の結果確認は、既存 `telemetry.log` ストリームに `info` ログを出力する方式で実現しなければならない。新規ログビューアを追加してはならない。

#### Scenario: 生成成功時にinfoログが出力される
- **WHEN** `PreparePrompts` が成功しリクエストが生成される
- **THEN** システムは `persona.requests.generated` の `info` ログを出力する
- **AND** ログには少なくとも `request_count`、`npc_count`、`task_id` を含める

#### Scenario: 生成失敗時にerrorログが出力される
- **WHEN** `PreparePrompts` 実行中にエラーが発生する
- **THEN** システムは `persona.requests.failed` の `error` ログを出力する
- **AND** ログには `task_id` と失敗理由を含める

