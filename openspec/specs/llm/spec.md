# LLM (LM Studio) 仕様書

## 1. 概要
本ドキュメントは `pkg/infrastructure/llm` のうち、**LM Studio プロバイダ実行に必要な契約とライフサイクル**を定義する。  
汎用インターフェース設計情報は本書に統合し、実装判断の一次情報を `spec.md` に一本化する。

## 2. 設計方針
- **Interface-First**: 呼び出し側は `LLMClient` 契約のみに依存し、HTTP実装詳細へ依存しない。
- **モデル明示指定**: モデル自動選択は許可しない。未指定はエラー。
- **ジョブ単位ライフサイクル**: `load -> requests -> unload` を1ジョブで一貫管理する。
- **Structured Output 優先**: JSON Schema 契約に基づく応答を標準経路とする。
- **後方互換**: 設定上の `local` / `local-llm` は `lmstudio` に正規化する。

## 3. 主要コンポーネント
- `LLMClient`: `ListModels`, `Complete`, `GenerateStructured`, `StreamComplete`, `GetEmbedding`, `HealthCheck` を提供する。
- `LLMManager`: `provider/model/endpoint/api_key` を解決し、`LLMClient` を返す。
- `ModelLifecycleClient`: `LoadModel`, `UnloadModel` によりジョブ単位のモデル状態を管理する。

## 4. Requirements

### Requirement: LLMモジュール共通のモデル一覧取得インターフェース
`pkg/infrastructure/llm` は、全プロバイダに対して共通の `ListModels(ctx)` 契約を提供しなければならない。呼び出し側はプロバイダ別のHTTP仕様を意識せず、正規化済みモデル情報を取得できなければならない。返却されるモデル情報は少なくとも `ID`、`DisplayName`、`MaxContextLength`、`Loaded` を含み、batch 実行可否のような provider 由来 capability を付加できなければならない。

#### Scenario: 共通契約でモデル一覧を取得する
- **WHEN** 呼び出し側が `LLMProvider.ListModels(ctx)` を実行する
- **THEN** プロバイダ実装は自プロバイダの一覧APIを呼び出して結果を取得する
- **AND** `ModelInfo` の共通形式で返却する

#### Scenario: batch 対応モデルを正規化して返す
- **WHEN** Gemini のように batch 対応モデルと非対応モデルが混在する provider から一覧を取得する
- **THEN** 実装は UI が利用できる batch 対応可否を正規化済みメタデータとして返さなければならない
- **AND** フロントエンドは provider 固有ルールをハードコードせずに mode 候補を判定できなければならない

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
`pkg/infrastructure/llm` は、設定で指定されたモデルを実行するために `ModelInfo` へ最低限 `ID`、`DisplayName`、`MaxContextLength`、`Loaded` を保持しなければならない。

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
`pkg/infrastructure/llm` の契約として Structured Output は全プロバイダサポートを要求し、LM Studio プロバイダが先行実装しなければならない。未実装プロバイダは `ErrStructuredOutputNotSupported` を返さなければならない。

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

### Requirement: LLM 実行は task 種別非依存で LM Studio 設定を解決しなければならない
`llm` は特定スライスの知識を持たず、Queue worker から渡される task 実行コンテキストに対して `config` から `provider/model` を再読込して実行しなければならない。`provider` は `lmstudio`、`gemini`、`xai` を受け付け、`local` / `local-llm` は `lmstudio` に正規化しなければならない。実行時は `sync` / `batch` の戦略とモデル capability を検証し、未対応の組み合わせでは外部実行を開始してはならない。

#### Scenario: 再開時に最新 config の provider と model で実行される
- **WHEN** Queue worker が request を再開する
- **THEN** `llm` は再開時点の `config` に保存された `provider` と `model` を使って実行経路を決定しなければならない
- **AND** モデル未指定時は実行を開始してはならない

#### Scenario: 未対応の batch 組み合わせは開始前に失敗する
- **WHEN** 呼び出し側が batch 非対応 provider または batch 非対応モデルで `batch` 実行を要求する
- **THEN** `llm` は provider / model unsupported エラーを返さなければならない
- **AND** 外部APIへの submit を開始してはならない

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

## 5. 参照資料
- [llm_test_spec.md](llm_test_spec.md)

## 6. ログ出力・テスト共通規約
1. テストは Table-Driven Test を基本とする。
2. Contract メソッドで Entry/Exit ログを出力する。
3. `context.Context` を伝播し、TraceID を全ログへ付与する。
4. 実行単位で `logs/*.jsonl` を出力する。
