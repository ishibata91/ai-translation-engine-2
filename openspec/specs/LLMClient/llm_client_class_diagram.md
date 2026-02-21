# LLMクライアント クラス図 (LLM Client Class Diagram)

```mermaid
classDiagram
    class LLMClient {
        <<interface>>
        +Complete(ctx Context, req Request) (Response, error)
        +StreamComplete(ctx Context, req Request) (StreamResponse, error)
        +GetEmbedding(ctx Context, text string) (List~float32~, error)
        +HealthCheck(ctx Context) error
    }

    class BatchClient {
        <<interface>>
        +SubmitBatch(ctx Context, reqs List~Request~) (BatchJobID, error)
        +GetBatchStatus(ctx Context, id BatchJobID) (BatchStatus, error)
        +GetBatchResults(ctx Context, id BatchJobID) (List~Response~, error)
    }

    class Request {
        +SystemPrompt string
        +UserPrompt string
        +MaxTokens int
        +Temperature float32
        +ResponseSchema Map~string, interface~
        +StopSequences List~string~
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
        +Next() (Response, bool)
        +Close() error
    }

    class GeminiProvider {
        -apiKey string
        -model string
        +Complete(...)
    }

    class OpenAIProvider {
        -apiKey string
        -endpoint string
        +Complete(...)
    }

    class LocalGGUFProvider {
        -modelPath string
        -serverURL string
        +Complete(...)
    }

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
    LLMClient <|.. OpenAIProvider
    LLMClient <|.. LocalGGUFProvider
    
    BatchClient <|.. OpenAIProvider : Optional
    
    LLMClient ..> Request : uses
    LLMClient ..> Response : returns
    Response *-- TokenUsage
```

## 主要エンティティの説明

### LLMClient (Interface)
全てのLLMプロバイダーが実装すべき基本インターフェース。同期的なリクエストと埋め込みベクトルの取得を定義する。

### BatchClient (Interface)
xAIやOpenAIのBatch APIのように、非同期で大量のリクエストを処理するプロバイダー向けのインターフェース。

### Request / Response (Structs)
リクエストパラメータとレスポンスデータをカプセル化する。`ResponseSchema`はJSON Schemaを保持し、モデルに構造化出力を強制するために使用する。

### Providers (Implementations)
各外部サービスやローカル実行エンジン（LM Studio等）への具体的な接続ロジックを持つ。これらは`LLMClient`インターフェースを満たすように実装される。
