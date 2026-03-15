# Master Persona Execution Flow 仕様

## Purpose
MasterPersona 実行を `request enqueue -> request execute -> persona save` の段階で追跡し、execution profile ベースで再開可能なフローを定義する。

## Requirements

### Requirement: MasterPersona は execution profile ベースの段階実行フローを提供しなければならない
システムは MasterPersona 実行において、`request enqueue` -> `request execute` -> `persona save` の 3 段階を単一 task ID で実行・追跡しなければならない。Master Persona は provider 名や batch API 名に直接依存せず、workflow / runtime が公開する execution profile を使って実行経路を決定しなければならない。`request execute` 段階では `sync` と `batch` の両方を扱え、phase 表示は provider 固有文言ではなく `REQUEST_ENQUEUED`、`REQUEST_EXECUTING_SYNC`、`BATCH_SUBMITTED`、`BATCH_POLLING`、`REQUEST_SAVING`、`REQUEST_COMPLETED` の共通 phase を使わなければならない。

#### Scenario: 未対応 execution profile は開始できない
- **WHEN** MasterPersona 開始時の execution profile が workflow / runtime capability 上で未対応である
- **THEN** システムはタスクを開始してはならない
- **AND** 無効な execution profile エラーを返さなければならない

#### Scenario: 段階実行が単一 task ID で追跡される
- **WHEN** MasterPersona を開始する
- **THEN** システムは enqueue、execute、save の各段階を同一 task ID に紐づけて状態更新しなければならない
- **AND** sync と batch の違いは共通 phase と message に反映されなければならない

### Requirement: MasterPersona の batch 再開は既存 batch job へ再接続しなければならない
MasterPersona が batch 実行中に停止または画面再接続された場合、再開処理は既存 batch job への再接続として扱わなければならない。`batch_job_id` を保持する request 群を新規 submit してはならない。batch API の詳細知識は runtime / gateway が保持し、Master Persona は再接続可能な execution profile であることだけを前提にしなければならない。

#### Scenario: batch submit 後に再開する
- **WHEN** task metadata または queue request が既存 `batch_job_id` を保持した状態で `ResumeTask` が呼ばれる
- **THEN** システムは provider へ既存 batch job の状態確認を行わなければならない
- **AND** 同一 request 群を二重投入してはならない

#### Scenario: batch 完了後に成功分を保存する
- **WHEN** batch 状態が `completed` または `partial_failed` になる
- **THEN** システムは取得可能な成功レスポンスを persona save 段階へ渡さなければならない
- **AND** 一部失敗があっても成功分の保存をスキップしてはならない

### Requirement: MasterPersona は task 終了時に中間生成物 cleanup を完了しなければならない
MasterPersona workflow は、task が終了したら `task_id` に紐づく persona 中間生成物 cleanup を persona slice 経由で完了しなければならない。workflow は artifact repository を直接参照してはならず、cleanup は必ず slice 契約を通じて実行しなければならない。

#### Scenario: 保存成功後に cleanup を実行する
- **WHEN** MasterPersona workflow が final persona の保存を完了して task を完了させる
- **THEN** システムは task 完了前に persona slice の cleanup 契約を呼び出さなければならない
- **AND** 当該 `task_id` の中間生成物を削除しなければならない

#### Scenario: 失敗または中止で終了する task でも cleanup を実行する
- **WHEN** MasterPersona task が failed または cancelled で終了する
- **THEN** システムは persona slice 経由で当該 `task_id` の中間生成物 cleanup を実行しなければならない
- **AND** workflow から artifact repository を直接呼び出してはならない

#### Scenario: cleanup は final 成果物を削除しない
- **WHEN** workflow が cleanup を実行する
- **THEN** システムは task スコープの中間生成物だけを削除しなければならない
- **AND** final 成果物や他 task の中間生成物を削除してはならない
