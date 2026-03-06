## MODIFIED Requirements

### Requirement: LLMジョブの再開時にモデル状態を復元できること
Queue は LLMジョブの再開時に、保存済みメタデータから同一プロバイダ・同一モデルを復元しなければならない。復元後はジョブ単位ロード/アンロードのライフサイクルを再適用しなければならない。

#### Scenario: 中断ジョブを同一モデルで再開する
- **WHEN** Queueが中断済みジョブを `provider` と `model` を含むメタデータで再開する
- **THEN** 再開処理は同一 `provider/model` でロードを実行する
- **AND** 再開後の結果は同一入力に対して互換な出力契約を満たす

### Requirement: LLMジョブメタデータの最小永続化項目
Queue は LLM再開性のために、`provider`、`model`、`request_fingerprint`、`structured_output_schema_version` をジョブメタデータとして保持しなければならない。

#### Scenario: 再開判定に必要なメタデータが揃っている
- **WHEN** Queueがジョブを永続化する
- **THEN** 上記4項目を欠損なく保存する
- **AND** 復元時に欠損がある場合はジョブを失敗として扱う

### Requirement: モデルロード失敗時は再試行しないこと
Queue とLM Studio連携では、モデルロード失敗を即時失敗として扱い、ロード再試行による滞留を発生させてはならない。

#### Scenario: load APIがエラーを返す
- **WHEN** ジョブ開始時のモデルロードが失敗する
- **THEN** Queueは当該実行を即時失敗にする
- **AND** load API の再試行は行わない
