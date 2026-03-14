## MODIFIED Requirements

### Requirement: Provider-native Batch API は共通の BatchClient 契約で扱わなければならない
`llm` は Gemini と xAI の provider-native Batch API を `BatchClient` 契約で扱わなければならない。`BatchClient` は `SubmitBatch`、`GetBatchStatus`、`GetBatchResults` を提供し、provider 固有レスポンスを上位層へ直接漏らしてはならない。さらに batch 結果の request 対応付けは配列順ではなく transport-level 相関 ID（`queue_job_id`）で復元しなければならない。相関 ID は LLM 本文ではなく request metadata / provider metadata 経路で往復させなければならない。

#### Scenario: provider 固有 state を共通 state へ正規化する
- **WHEN** `BatchClient.GetBatchStatus` が Gemini または xAI の batch 状態を取得する
- **THEN** 実装は `queued`、`running`、`completed`、`partial_failed`、`failed`、`cancelled` の共通状態へ正規化して返さなければならない
- **AND** workflow と UI は provider 固有状態名に依存してはならない

#### Scenario: 部分失敗でも結果取得を継続する
- **WHEN** batch 実行で一部 request が成功し、一部 request が失敗する
- **THEN** `BatchClient` は `partial_failed` を返さなければならない
- **AND** 呼び出し側は成功分の結果取得と保存を継続できなければならない

#### Scenario: xAI は batch_request_id で相関 ID を往復する
- **WHEN** xAI Batch へ request を submit する
- **THEN** 実装は `batch_request_id` に `queue_job_id` を設定して送信しなければならない
- **AND** `GetBatchResults` では `batch_request_id` を `Response.Metadata.queue_job_id` として返さなければならない

#### Scenario: Gemini は inlined metadata で相関 ID を往復する
- **WHEN** Gemini Batch へ inlined requests を submit する
- **THEN** 実装は `inlinedRequest.metadata.queue_job_id` を送信しなければならない
- **AND** `GetBatchResults` では `inlinedResponse.metadata.queue_job_id` を `Response.Metadata.queue_job_id` として返さなければならない

#### Scenario: 相関は LLM 出力本文に依存しない
- **WHEN** batch の応答本文が空または解析不能でも provider metadata が取得できる
- **THEN** 実装は `queue_job_id` による request 対応付けを継続しなければならない
- **AND** prompt/response テキスト内容による識別を行ってはならない