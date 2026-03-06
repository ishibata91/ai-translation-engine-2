# Persona Request Preview 仕様書

## 概要
`MasterPersona` 画面からペルソナ生成Phase 1（リクエスト生成）を起動し、タスク状態と既存ログビューアで結果確認できるようにする機能。

## ADDED Requirements

### Requirement: MasterPersona開始操作はペルソナリクエスト生成タスクを起動しなければならない
`MasterPersona` 画面の開始ボタンは、`task.Bridge` 経由で「マスターペルソナ生成タスク」を起動しなければならない。システムはタスクIDを発行し、同一IDで進捗・完了・失敗を追跡可能でなければならない。

#### Scenario: 開始ボタン押下でタスクが起動される
- **WHEN** ユーザーが `MasterPersona` 画面で開始ボタンを押下する
- **THEN** システムは `task.Bridge` に「マスターペルソナ生成タスク」の開始を要求する
- **AND** タスクIDを発行して実行状態を `Running` に更新する

#### Scenario: タスクが完了しUI状態が更新される
- **WHEN** `PreparePrompts` が正常終了する
- **THEN** システムはタスクステータスを `REQUEST_GENERATED` に更新する
- **AND** UIは生成中表示を解除し、結果サマリを表示可能な状態にする

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

