# Spec: progress-notifier

## Overview

`pkg/infrastructure/progress` パッケージとして提供される汎用進捗通知インフラ。ドメイン知識を一切持たず、任意のスライス・インフラコンポーネントが DI 経由で利用できる。

## Requirements

### Requirement: Generic ProgressEvent

```go
// ProgressEvent はドメイン・インフラに依存しない汎用進捗イベント型。
type ProgressEvent struct {
    CorrelationID string // 処理のまとまりを識別するID (UUID等)。UI側での表示単位となる
    Total         int    // 総件数（不明な場合は 0）
    Completed     int    // 完了件数
    Failed        int    // 失敗件数
    Status        string // "IN_PROGRESS" / "COMPLETED" / "FAILED"
    Message       string // ユーザーに表示する進捗メッセージ
}
```

- `CorrelationID` は、特定のジョブIDではなく、エンドユーザーから見た「ひとつの処理（翻訳プロセス等）」を識別する。
- `Status` は文字列型とし、定数 `StatusInProgress`, `StatusCompleted`, `StatusFailed` を同パッケージ内で定義する。
- `Total = 0` の場合、UI 側は総件数不明として扱う（不定量ストリーミング対応）。

### Requirement: ProgressNotifier Interface

```go
// ProgressNotifier は進捗通知の送信先を抽象化するインターフェース。
// 実装は WebSocket / SSE / ログ出力 / テスト用モック 等を想定。
type ProgressNotifier interface {
    OnProgress(ctx context.Context, event ProgressEvent)
}
```

- 引数に `ctx context.Context` を必ず受け取り、OpenTelemetry TraceID を伝播可能にする。
- `OnProgress` はノンブロッキングを保証しない（実装者の責務）。呼び出し元はゴルーチン内から呼び出すことを想定する。
- インターフェースのみを定義し、具象実装（WebSocket 送信等）はこのパッケージに含めない。

### Requirement: NoopNotifier

テスト・開発環境向けの何もしない実装を同パッケージ内に提供する。

```go
// NoopNotifier は OnProgress を何もせずに無視するデフォルト実装。
// テストや、進捗通知が不要なシナリオで ProgressNotifier の代替として使用する。
type NoopNotifier struct{}

func (n *NoopNotifier) OnProgress(_ context.Context, _ ProgressEvent) {}
```

### Requirement: Wire Provider

`google/wire` で DI 解決できるよう、`NoopNotifier` の Provider 関数を公開する。

```go
// NewNoopNotifier は NoopNotifier を ProgressNotifier として返す Wire Provider。
func NewNoopNotifier() ProgressNotifier {
    return &NoopNotifier{}
}

// ProviderSet は progress パッケージの Wire ProviderSet。
var ProviderSet = wire.NewSet(NewNoopNotifier)
```

### Requirement: Context Propagation (refactoring_strategy.md §7 準拠)

- `OnProgress` に渡される `ctx` は、呼び出し元（ワーカー等）が伝播させた OpenTelemetry トレースコンテキストを含む。
- 具象実装は `slog.DebugContext(ctx, "OnProgress", ...)` 等でトレースIDをログに記録することが推奨される。

## Scenarios

#### Scenario: 処理中の進捗通知
- **WHEN** 任意のコンポーネントが処理のステップを1つ完了した
- **THEN** `ProgressNotifier.OnProgress` が `Status = "IN_PROGRESS"`, `Completed = n` で呼び出される

#### Scenario: 処理完了通知
- **WHEN** 全行程が完了した
- **THEN** `ProgressNotifier.OnProgress` が `Status = "COMPLETED"`, `Completed == Total` で呼び出される

#### Scenario: 異常終了の通知
- **WHEN** 処理が途中で失敗した
- **THEN** `ProgressNotifier.OnProgress` が `Status = "FAILED"`, `Failed > 0` かつ `Message` にエラー情報を伴って呼び出される

#### Scenario: NoopNotifier がパニックしない
- **WHEN** `NoopNotifier.OnProgress` を任意の `ProgressEvent` で呼び出す
- **THEN** パニックせず正常に戻る（何もしない）

#### Scenario: Total = 0 での通知
- **WHEN** 総件数不明のまま `Total = 0` で `OnProgress` を呼び出す
- **THEN** UI側エラーなく受け入れられ、件数表示は「不明」として扱われる
