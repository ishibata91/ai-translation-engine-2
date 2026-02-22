# LLMクライアント シーケンス図 (LLM Client Sequence Diagram)

## 1. 同期リクエスト (Complete)
通常の翻訳処理などで使用される。

```mermaid
sequenceDiagram
    participant App as 呼び出し元 (Slice Logic)
    participant Manager as LLMManager
    participant Client as LLMClient (Interface)
    participant Provider as Provider Implementation (e.g. Gemini)
    participant API as External LLM API

    App->>Manager: GetClient(ctx, LLMConfig)
    Manager-->>App: Client
    App->>Client: Complete(ctx, Request)
    Client->>Provider: Complete(ctx, Request)
    Provider->>Provider: buildRequest() [private]
    Provider->>API: POST /chat/completions (HTTP)
    API-->>Provider: APIレスポンス (JSON)
    Provider->>Provider: parseResponse() [private]
    Provider-->>Client: Response
    Client-->>App: Response
```

## 2. 非同期バッチリクエスト — Gemini (Batch API)
Gemini向けバッチ処理（OpenAI互換フロー）。

```mermaid
sequenceDiagram
    participant App as 呼び出し元 (Slice Logic)
    participant Manager as LLMManager
    participant Batch as BatchClient (Gemini)
    participant API as Gemini Batch API

    App->>Manager: GetBatchClient(ctx, LLMConfig{Provider:"gemini"})
    Manager-->>App: BatchClient
    App->>Batch: SubmitBatch(ctx, []Request)
    Batch->>API: POST バッチジョブ作成
    API-->>Batch: BatchJobID
    Batch-->>App: BatchJobID

    App->>Batch: GetBatchStatus(ctx, BatchJobID)
    Batch->>API: GET ステータス確認
    API-->>Batch: Status (completed)
    Batch-->>App: BatchStatus

    App->>Batch: GetBatchResults(ctx, BatchJobID)
    Batch->>API: 結果取得
    API-->>Batch: []Response
    Batch-->>App: []Response
```

## 3. 非同期バッチリクエスト — xAI 独自フロー (BatchClient)
xAI Batch API は OpenAI Batch API と**非互換**の独自フォーマットを使用する。

```mermaid
sequenceDiagram
    participant App as 呼び出し元 (Slice Logic)
    participant Manager as LLMManager
    participant Batch as xAIBatchClient
    participant API as xAI Batch API (api.x.ai/v1)

    Note over App, API: ① バッチジョブ作成
    App->>Manager: GetBatchClient(ctx, LLMConfig{Provider:"xai"})
    Manager-->>App: xAIBatchClient
    App->>Batch: SubmitBatch(ctx, []Request)
    Batch->>Batch: _createBatch(name) [private]
    Batch->>API: POST /v1/batches {name, endpoint, completion_window}
    API-->>Batch: {"batch_id": "..."}  ※ id ではなく batch_id
    Batch->>Batch: _addRequests(batchID, reqs) [private]
    Batch->>API: POST /v1/batches/{batch_id}/requests
    Note right of API: {batch_requests:[{batch_request_id,<br/>batch_request:{chat_get_completion:{...}}}]}
    API-->>Batch: 200 OK
    Batch->>Batch: _pollUntilCompleted(batchID) [private]

    Note over App, API: ② ポーリング (state.num_pending > 0 の間繰り返す)
    loop ポーリング (30秒間隔)
        Batch->>API: GET /v1/batches/{batch_id}
        API-->>Batch: {state: {num_requests, num_pending, num_success, num_error}}
    end
    API-->>Batch: num_pending == 0 → completed
    Batch-->>App: BatchJobID

    Note over App, API: ③ 結果取得 (ページネーション)
    App->>Batch: GetBatchResults(ctx, BatchJobID)
    Batch->>Batch: _parseResults(data) [private]
    loop pagination_token が存在する間
        Batch->>API: GET /v1/batches/{batch_id}/results?pagination_token=...
        API-->>Batch: {results: [...], pagination_token: "..."}
    end
    Note right of Batch: batch_result.response<br/>.chat_get_completion.choices[0]<br/>.message.content を抽出
    Batch-->>App: []Response
```

> **対応モデル**: `grok-3`, `grok-4-*` のみ。`grok-3-mini` は Batch API 非対応。

## 4. ストリーミングリクエスト (StreamComplete)
UIでリアルタイムに生成過程を表示する場合に使用される。

```mermaid
sequenceDiagram
    participant UI as React Frontend
    participant Srv as Go Server (API)
    participant Manager as LLMManager
    participant Client as LLMClient
    participant Provider as Provider

    UI->>Srv: WebSocket / SSE Request
    Srv->>Manager: GetClient(ctx, LLMConfig)
    Manager-->>Srv: Client
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
