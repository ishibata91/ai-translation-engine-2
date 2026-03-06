## ADDED Requirements

### Requirement: LLMモジュール共通のモデル一覧取得インターフェース
`pkg/llm` は、全プロバイダに対して共通の `ListModels(ctx)` 契約を提供しなければならない。呼び出し側はプロバイダ別のHTTP仕様を意識せず、正規化済みモデル情報を取得できなければならない。

#### Scenario: 共通契約でモデル一覧を取得する
- **WHEN** 呼び出し側が `LLMProvider.ListModels(ctx)` を実行する
- **THEN** プロバイダ実装は自プロバイダの一覧APIを呼び出して結果を取得する
- **AND** `ModelInfo` の共通形式で返却する

### Requirement: LM Studioモデル一覧API仕様を明記する
LM Studio プロバイダは、モデル一覧取得において `GET /api/v1/models` を使用しなければならない。リクエストとレスポンス仕様は以下に固定する。

API仕様:
- Method: `GET`
- Path: `/api/v1/models`
- Headers:
  - `Authorization: Bearer <LM_API_TOKEN>`（設定されている場合）
- Success Response: `200 OK`
- Response Body（抜粋）:
  - `models[]`
    - `type` (`llm` / `embedding`)
    - `key`
    - `display_name`
    - `loaded_instances[]`
      - `id`
      - `config.context_length`
    - `max_context_length`

#### Scenario: LM Studioレスポンスを共通モデルへ正規化する
- **WHEN** `GET /api/v1/models` が `200` と `models[]` を返す
- **THEN** 実装は `type=llm` を推論候補として抽出する
- **AND** `key` を内部 `ModelInfo.ID` に、`display_name` を `ModelInfo.DisplayName` に正規化する

### Requirement: 指定モデル実行に必要な最小メタデータを保持する
`pkg/llm` は、設定で指定されたモデルを実行するために `ModelInfo` へ最低限 `ID`、`DisplayName`、`MaxContextLength`、`Loaded` を保持しなければならない。

#### Scenario: モデル未指定はエラーにする
- **WHEN** 呼び出し側がモデル名を明示しない
- **THEN** 実装はモデル自動選択を行わず、モデル未指定エラーを返す
- **AND** 呼び出し側は `pkg/config` 経由で指定したモデルIDを保存し、実行時にその値を渡す

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
