# Design: Progress Notifier

## Context

`job-queue-infrastructure` の設計過程で `ProgressNotifier` インターフェースの必要性が確認された。しかし JobQueue パッケージ内に定義するとVSAの依存方向に違反するため、`pkg/infrastructure/progress` として Shared Infrastructure レイヤーに独立切り出しする。

`refactoring_strategy.md §5` の Shared Kernel 判断基準：「技術的関心事（Infrastructure/Utility レベル）」→ `Shared/Core レイヤーへ DRY に集約` が適用される。

## Goals / Non-Goals

**Goals:**
- ドメイン非依存の `ProgressEvent` 型と `ProgressNotifier` インターフェースを `pkg/infrastructure/progress` に定義する。
- テスト・開発環境向け `NoopNotifier` をデフォルト実装として提供する。
- `google/wire` の `ProviderSet` を公開し、任意のスライス・インフラが DI で利用できるようにする。

**Non-Goals:**
- WebSocket / SSE などの具体的な通知実装はこのパッケージに含めない（ProcessManager 側の責務）。
- `ProgressEvent` にドメイン固有のフィールドを追加すること。

## Decisions

### 1. パッケージ配置

```
pkg/
└── infrastructure/
    └── progress/
        ├── notifier.go   ← ProgressEvent 型・ProgressNotifier インターフェース・定数
        ├── noop.go       ← NoopNotifier 実装
        └── provider.go   ← Wire ProviderSet
```

`pkg/infrastructure/` 以下に配置することで、技術的関心事（Shared Infrastructure）として明確に分類する。ドメインスライス（`pkg/term_translator_slice/` 等）は `pkg/infrastructure/progress` にのみ依存し、循環依存は発生しない。

### 2. インターフェース設計

```go
// pkg/infrastructure/progress/notifier.go

package progress

import "context"

const (
    StatusInProgress = "IN_PROGRESS"
    StatusCompleted  = "COMPLETED"
    StatusFailed     = "FAILED"
)

type ProgressEvent struct {
    CorrelationID string
    Total         int    // 0 = 総件数不明
    Completed     int
    Failed        int
    Status        string // StatusInProgress / StatusCompleted / StatusFailed
    Message       string
}

type ProgressNotifier interface {
    OnProgress(ctx context.Context, event ProgressEvent)
}
```

- `Status` は文字列定数とし、enum ライクに管理する。型エイリアスは使用しない（シリアライズ互換性を優先）。
- `context.Context` を第一引数に受け取り、OpenTelemetry TraceID が全ての `OnProgress` 呼び出しに伝播する。

### 3. NoopNotifier と Wire Provider

```go
// pkg/infrastructure/progress/noop.go
type NoopNotifier struct{}
func (n *NoopNotifier) OnProgress(_ context.Context, _ ProgressEvent) {}

// pkg/infrastructure/progress/provider.go
var ProviderSet = wire.NewSet(NewNoopNotifier)
func NewNoopNotifier() ProgressNotifier { return &NoopNotifier{} }
```

- `ProviderSet` は `NoopNotifier` を `ProgressNotifier` インターフェースとしてバインドする。
- ProcessManager が WebSocket 実装を提供する際は、`wire.Bind` で差し替える。

### 4. 利用側パターン

どのようなコンポーネントでも、`ProgressNotifier` を DI で受け取るだけで進捗報告が可能です。

#### A. ドメインスライスの例（逐次処理）
```go
type contextEngineSlice struct {
    notifier progress.ProgressNotifier
}

func (s *contextEngineSlice) BuildContext(ctx context.Context, req BuildRequest) {
    for i, item := range items {
        // ...処理...
        s.notifier.OnProgress(ctx, progress.ProgressEvent{
            CorrelationID: req.ProcessID, // UIが発行した実行IDを渡す
            Total:         len(items),
            Completed:     i + 1,
            Status:        progress.StatusInProgress,
            Message:       fmt.Sprintf("Item %d processed", i+1),
        })
    }
}
```

#### B. インフラワーカーの例（非同期処理）
```go
type worker struct {
    notifier progress.ProgressNotifier
}

func (w *worker) run(ctx context.Context, job Job) {
    // LLM Batch API のポーリング結果などを流す
    w.notifier.OnProgress(ctx, progress.ProgressEvent{
        CorrelationID: job.ExternalID,
        Status:        progress.StatusInProgress,
        Message:       "Waiting for LLM batch results...",
    })
}
```

### 5. 構造化ログ (refactoring_strategy.md §6.2/§7 準拠)

- `NoopNotifier` はログを出力しない（何もしない）。
- 将来の具象実装（WebSocket Notifier 等）は `slog.DebugContext(ctx, "OnProgress called", ...)` で Entry ログを出力し、TraceID を含む構造化ログを記録することを推奨する。

## Risks / Trade-offs

- **[Trade-off] `OnProgress` のブロッキング**: インターフェースはブロッキング・ノンブロッキングを規定しない。WebSocket 送信等でブロックしないよう、実装側がゴルーチンまたはバッファ付きチャネルを管理する責務を負う。
- **[Trade-off] エラーハンドリング**: `OnProgress` は `error` を返さない。通知失敗は処理フローを中断せず、必要に応じて実装側がログで記録する設計とする。
