# ワークフロー コア

controller から usecase slice と runtime を接続する workflow capability の責務と要件を定義する。

## Requirements

### Requirement: Workflow は controller の唯一のユースケース入口でなければならない
controller は usecase slice や runtime を直接呼び出さず、workflow 契約だけを呼び出さなければならない。

#### Scenario: controller は workflow だけを呼び出す
- **WHEN** UI がユースケース開始または再開を要求する
- **THEN** controller は workflow 契約だけを呼び出さなければならない

### Requirement: Workflow は runtime を通じて実行制御を行わなければならない
workflow は queue、progress、workflow state などの実行制御を runtime 契約経由で行わなければならない。

#### Scenario: workflow は runtime queue を利用する
- **WHEN** workflow が request を enqueue または resume する
- **THEN** workflow は runtime queue 契約を利用しなければならない

### Requirement: Workflow は provider と実行モードに応じて runtime 実行経路を決定しなければならない
workflow は controller から受けたユースケース要求に対し、選択中 provider と実行モードに応じて runtime へ `sync` または `batch` の実行を委譲しなければならない。workflow 自身が provider 固有 HTTP 実装や batch polling を持ってはならない。

#### Scenario: workflow は batch 実行を runtime へ委譲する
- **WHEN** MasterPersona の実行モードが `batch` として保存されている
- **THEN** workflow は runtime queue に batch 実行前提の resume / 実行を委譲しなければならない
- **AND** workflow は provider 固有 API を直接呼び出してはならない

#### Scenario: workflow は未対応の provider と mode を開始前に拒否する
- **WHEN** provider と実行モードの組み合わせが capability 上未対応である
- **THEN** workflow は runtime 実行前にエラーを返さなければならない
- **AND** task の phase を進めてはならない

### Requirement: Workflow は batch submit 後の状態追跡と再開を統一 phase で管理しなければならない
workflow は batch submit 後の状態追跡を `BATCH_SUBMITTED`、`BATCH_POLLING`、`REQUEST_SAVING`、`REQUEST_COMPLETED` の provider 非依存 phase で管理しなければならない。既存 batch job を再開する場合は再 submit ではなく既存 job への再接続として扱わなければならない。

#### Scenario: batch submit 済み task を再開する
- **WHEN** workflow が `batch_job_id` を保持する task を再開する
- **THEN** workflow は既存 batch job の状態確認から再開しなければならない
- **AND** 同一 request 群を新規 submit してはならない

#### Scenario: batch 実行の進捗を provider 非依存 phase で通知する
- **WHEN** runtime が batch の状態更新を返す
- **THEN** workflow は provider 固有文言ではなく共通 phase と進捗イベントで UI へ通知しなければならない
- **AND** UI は `lmstudio dispatch` のような provider 固有 phase 名へ依存してはならない
