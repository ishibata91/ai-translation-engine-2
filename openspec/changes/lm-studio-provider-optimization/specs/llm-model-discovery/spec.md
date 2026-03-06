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
