# LLMClient Delta Spec

## ADDED Requirements

### Requirement: 同期バルクリクエストの処理 (CompleteBulk)
LLMClient（または共通ヘルパー関数）は、複数の `Request` を受け取り、指定された `Concurrency` に基づいて並列に処理を行い、すべてのリクエストが完了した時点で `[]Response` を返さなければならない。

#### Scenario: 正常なバルク処理
- **WHEN** 10件のリクエストリストと `Concurrency=3` が渡された
- **THEN** ワーカープールが構築され、最大3並列でAPI呼び出しが行われる
- **AND** すべての処理が完了すると要素数10のレスポンスリストが返る

#### Scenario: 一部のリクエストが失敗した場合 (Partial Failure)
- **WHEN** 複数件のバルクリクエストのうち、1件がAPIエラー（レートリミット等）になった
- **THEN** 成功したリクエストは正常な `Response` が返り、失敗したリクエストのみ `Response.Success = false` と `Response.Error` が格納される
- **AND** 全体の関数呼び出しとしてはエラーを返さず処理を完了する

### Requirement: UIコンフィグによる並列数の適用
`LLMManager` はプロバイダごとのクライアントを初期化する際、またはバルク処理ヘルパーを実行する際に、`ConfigStore` から取得した並列数（`Concurrency`）を適用しなければならない。

#### Scenario: 並列数の動的適用
- **WHEN** ユーザーが設定画面から Gemini の並列実行数を `5` に設定した
- **THEN** 以降のバルク同期リクエスト処理では `Concurrency=5` として並列実行される
- **AND** ロックやチャネル競合なく安全に流量制御される
