# LLMクライアント シーケンス図 (LLM Client Sequence Diagram)

## 1. 同期リクエスト (Complete)
通常の翻訳処理などで使用される。

```mermaid
sequenceDiagram
    participant App as 呼び出し元 (Slice Logic)
    participant Client as LLMClient (Interface)
    participant Provider as Provider Implementation (e.g. Gemini)
    participant API as External LLM API

    App->>Client: Complete(ctx, Request)
    Client->>Provider: Complete(ctx, Request)
    Provider->>Provider: リクエストの変換 (Go Struct -> API JSON)
    Provider->>API: POST /chat/completions (HTTP/gRPC)
    API-->>Provider: APIレスポンス (JSON)
    Provider->>Provider: レスポンスの変換 (API JSON -> Go Struct)
    Provider-->>Client: Response
    Client-->>App: Response
```

## 2. 非同期バッチリクエスト (Batch API)
大規模な初期翻訳や、コスト削減のために非同期処理を行う場合に使用される。

```mermaid
sequenceDiagram
    participant App as 呼び出し元 (Slice Logic)
    participant Batch as BatchClient (Interface)
    participant Provider as Provider (e.g. OpenAI)
    participant API as Batch API
    participant Store as Job Store (DB/File)

    Note over App, API: ジョブの投入
    App->>Batch: SubmitBatch(ctx, []Request)
    Batch->>Provider: SubmitBatch(ctx, []Request)
    Provider->>API: Create Batch Job
    API-->>Provider: BatchID
    Provider->>Store: 保存 (BatchID, Provider, Status="pending")
    Provider-->>Batch: BatchJobID
    Batch-->>App: BatchJobID

    Note over App, API: ステータス確認 (ポーリングまたはタイマー)
    App->>Batch: GetBatchStatus(ctx, BatchJobID)
    Batch->>Provider: GetBatchStatus(ctx, BatchJobID)
    Provider->>API: Retrieve Batch Status
    API-->>Provider: Status (completed)
    Provider-->>Batch: BatchStatus
    Batch-->>App: BatchStatus

    Note over App, API: 結果取得
    App->>Batch: GetBatchResults(ctx, BatchJobID)
    Batch->>Provider: GetBatchResults(ctx, BatchJobID)
    Provider->>API: Download Results File
    API-->>Provider: Result Data (JSONL)
    Provider->>Provider: パース
    Provider-->>Batch: []Response
    Batch-->>App: []Response
```

## 3. ストリーミングリクエスト (StreamComplete)
UIでリアルタイムに生成過程を表示する場合に使用される。

```mermaid
sequenceDiagram
    participant UI as React Frontend
    participant Srv as Go Server (API)
    participant Client as LLMClient
    participant Provider as Provider

    UI->>Srv: WebSocket / SSE Request
    Srv->>Client: StreamComplete(ctx, Request)
    Client->>Provider: StreamComplete(ctx, Request)
    Provider->>Provider: HTTP Stream開始
    loop 各トークン受信
        Provider-->>Client: Partial Response
        Client-->>Srv: Partial Response
        Srv-->>UI: Send Chunk
    end
    Provider-->>Client: Close
    Client-->>Srv: Close
    Srv-->>UI: End Stream
```
