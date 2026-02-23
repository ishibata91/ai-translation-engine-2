# LLMクライアント クラス図 (LLM Client Class Diagram)

```mermaid
classDiagram
    direction TB

    class LLMClient {
        <<interface>>
        +Complete(ctx Context, req Request) Response, error
        +StreamComplete(ctx Context, req Request) StreamResponse, error
        +GetEmbedding(ctx Context, text string) []float32, error
        +HealthCheck(ctx Context) error
    }

    class LLMManager {
        <<interface>>
        +GetClient(ctx Context, config LLMConfig) LLMClient, error
        +GetBatchClient(ctx Context, config LLMConfig) BatchClient, error
    }

    class LLMConfig {
        +Provider string
        +APIKey string
        +Endpoint string
        +Model string
        +Parameters Map~string,interface~
    }

    class BatchClient {
        <<interface>>
        +SubmitBatch(ctx Context, reqs []Request) BatchJobID, error
        +GetBatchStatus(ctx Context, id BatchJobID) BatchStatus, error
        +GetBatchResults(ctx Context, id BatchJobID) []Response, error
    }

    class Request {
        +SystemPrompt string
        +UserPrompt string
        +MaxTokens int
        +Temperature float32
        +ResponseSchema Map~string,interface~
        +StopSequences []string
    }

    class Response {
        +Content string
        +Success bool
        +Error string
        +Usage TokenUsage
    }

    class TokenUsage {
        +PromptTokens int
        +CompletionTokens int
        +TotalTokens int
    }

    class StreamResponse {
        <<interface>>
        +Next() Response, bool
        +Close() error
    }

    class GeminiProvider {
        -apiKey string
        -model string
        -logger slog.Logger
        +Complete(...)
    }

    class LocalProvider {
        -serverURL string
        -logger slog.Logger
        +Complete(...)
    }

    class xAIProvider {
        -apiKey string
        -model string
        -logger slog.Logger
        +Complete(...)
    }

    class xAIBatchClient {
        -apiKey string
        -model string
        -logger slog.Logger
        +SubmitBatch(ctx, reqs) BatchJobID, error
        +GetBatchStatus(ctx, id) BatchStatus, error
        +GetBatchResults(ctx, id) []Response, error
        -createBatch(name) string
        -addRequests(batchID, reqs) error
        -pollUntilCompleted(batchID) error
        -parseResults(data) []Response
    }

    note for xAIBatchClient "xAI独自フォーマット(OpenAI Batch非互換)\nPOST /v1/batches → batch_id\nPOST /v1/batches/{id}/requests\nGET /v1/batches/{id} (state.*)\nGET /v1/batches/{id}/results (paged)\n⚠ grok-3-mini は非対応"

    class BatchJobID {
        +ID string
        +Provider string
    }

    class BatchStatus {
        +ID string
        +State string
        +Progress float32
    }

    LLMClient <|.. GeminiProvider
    LLMClient <|.. LocalProvider
    LLMClient <|.. xAIProvider

    BatchClient <|.. xAIBatchClient
    BatchClient <|.. GeminiProvider : Optional

    LLMClient ..> Request : uses
    LLMClient ..> Response : returns
    Response *-- TokenUsage

    LLMManager ..> LLMClient : creates
    LLMManager ..> BatchClient : creates
    LLMManager ..> LLMConfig : uses

    xAIProvider --> xAIBatchClient : creates
```

## 主要エンティティの説明

### LLMClient (Interface)
全てのLLMプロバイダーが実装すべき基本インターフェース。同期的なリクエストと埋め込みベクトルの取得を定義する。

### BatchClient (Interface)
xAIやGeminiのBatch APIのように、非同期で大量のリクエストを処理するプロバイダー向けのインターフェース。

### xAIBatchClient — xAI 独自バッチ実装

xAI Batch API は **OpenAI Batch API と非互換**の独自フォーマットを採用する。

| 項目                 | xAI 独自仕様                                                            |
| -------------------- | ----------------------------------------------------------------------- |
| バッチ作成レスポンス | `batch_id`（OpenAI は `id`）                                            |
| リクエスト追加       | `POST /v1/batches/{id}/requests`（別途追加が必要）                      |
| リクエスト形式       | `batch_requests[].batch_request.chat_get_completion`                    |
| ステータス           | `state.{num_requests, num_pending, num_success, num_error}` から導出    |
| 結果取得             | `GET /v1/batches/{id}/results`（`pagination_token` でページネーション） |
| 結果パス             | `batch_result.response.chat_get_completion.choices[0].message.content`  |
| Batch対応モデル      | `grok-3`, `grok-4-*`（`grok-3-mini` は**非対応**）                      |

プライベートメソッドは同一ファイル内に分割（SRP原則）: `_createBatch`, `_addRequests`, `_pollUntilCompleted`, `_parseResults`

### Request / Response (Structs)
リクエストパラメータとレスポンスデータをカプセル化する。`ResponseSchema`はJSON Schemaを保持し、モデルに構造化出力を強制するために使用する。

### Providers (Implementations)
各外部サービスやローカル実行エンジンへの具体的な接続ロジックを持つ。これらは`LLMClient`インターフェースを満たすように実装される。
