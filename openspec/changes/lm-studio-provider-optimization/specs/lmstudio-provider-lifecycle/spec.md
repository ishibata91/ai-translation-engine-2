## ADDED Requirements

### Requirement: LM StudioロードAPIをジョブ開始時に呼び出す
LM Studio プロバイダは、Queueジョブ開始時に `POST /api/v1/models/load` を呼び出して対象モデルをロードしなければならない。レスポンスの `instance_id` はジョブ内コンテキストに保持し、アンロードで再利用しなければならない。

API仕様:
- Method: `POST`
- Path: `/api/v1/models/load`
- Headers:
  - `Authorization: Bearer <LM_API_TOKEN>`（設定されている場合）
  - `Content-Type: application/json`
- Request Body:
  - `model` (string, required)
  - `context_length` (number, optional)
  - `flash_attention` (boolean, optional)
  - `echo_load_config` (boolean, optional)
- Success Response: `200 OK`
- Response Body（抜粋）:
  - `type`
  - `instance_id`
  - `status` (`loaded`)
  - `load_config`

#### Scenario: ジョブ開始時にモデルをロードする
- **WHEN** QueueワーカーがLM Studioプロバイダでジョブ実行を開始する
- **THEN** `POST /api/v1/models/load` を1回実行して `instance_id` を取得する
- **AND** 取得失敗時は即時エラー終了し、load再試行は行わない

### Requirement: LM StudioアンロードAPIをジョブ終了時と中断時に呼び出す
LM Studio プロバイダは、ジョブ完了時またはキャンセル時に `POST /api/v1/models/unload` を呼び出してモデルを解放しなければならない。

API仕様:
- Method: `POST`
- Path: `/api/v1/models/unload`
- Headers:
  - `Authorization: Bearer <LM_API_TOKEN>`（設定されている場合）
  - `Content-Type: application/json`
- Request Body:
  - `instance_id` (string, required)
- Success Response: `200 OK`
- Response Body（抜粋）:
  - `instance_id`

#### Scenario: ジョブ完了時にアンロードする
- **WHEN** ジョブ内の全リクエスト処理が成功または失敗で終了する
- **THEN** `POST /api/v1/models/unload` を1回実行する
- **AND** リクエスト単位ではアンロードしない

#### Scenario: ジョブ中断時にアンロードする
- **WHEN** ジョブ実行中に `context cancellation` が発生する
- **THEN** `POST /api/v1/models/unload` を1回実行する
- **AND** アンロード実行後にキャンセルエラーを返す

### Requirement: Structured Output契約をLM Studioで先行実装する
`pkg/llm` の契約として Structured Output は全プロバイダサポートを要求し、今回のチェンジでは LM Studio プロバイダが先行実装しなければならない。未実装プロバイダは `ErrStructuredOutputNotSupported` を返さなければならない。

API仕様 (OpenAI互換):
- Method: `POST`
- Path: `/v1/chat/completions`
- Headers:
  - `Content-Type: application/json`
  - `Authorization: Bearer <LM_API_TOKEN>`（設定されている場合）
- Request Body（必須要素）:
  - `model`
  - `messages[]`
  - `response_format.type = "json_schema"`
  - `response_format.json_schema.name`
  - `response_format.json_schema.strict = true` (boolean)
  - `response_format.json_schema.schema` (JSON Schema object)
- Success Response: `200 OK` with structured JSON content

#### Scenario: JSON Schema付きの構造化応答を取得する
- **WHEN** 呼び出し側が `GenerateStructured` へJSON Schemaを渡す
- **THEN** 実装は `/v1/chat/completions` に `response_format.json_schema` を付与して送信する
- **AND** 応答をスキーマ契約に従って検証して返却する
