## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Provider-native Batch API は共通の BatchClient 契約で扱わなければならない
`llm` は Gemini と xAI の provider-native Batch API を `BatchClient` 契約で扱わなければならない。`BatchClient` は `SubmitBatch`、`GetBatchStatus`、`GetBatchResults` を提供し、provider 固有レスポンスを上位層へ直接漏らしてはならない。

#### Scenario: provider 固有 state を共通 state へ正規化する
- **WHEN** `BatchClient.GetBatchStatus` が Gemini または xAI の batch 状態を取得する
- **THEN** 実装は `queued`、`running`、`completed`、`partial_failed`、`failed`、`cancelled` の共通状態へ正規化して返さなければならない
- **AND** workflow と UI は provider 固有状態名に依存してはならない

#### Scenario: 部分失敗でも結果取得を継続する
- **WHEN** batch 実行で一部 request が成功し、一部 request が失敗する
- **THEN** `BatchClient` は `partial_failed` を返さなければならない
- **AND** 呼び出し側は成功分の結果取得と保存を継続できなければならない
