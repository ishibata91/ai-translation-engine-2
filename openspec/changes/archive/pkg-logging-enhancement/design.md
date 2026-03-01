# Design: pkg 全体へのログガイド第1部適用

## 1. ログ共通フォーマット
[`specs/log-guide.md`](specs/log-guide.md) の第3部に基づき、以下のキーを標準的に含める。

| キー | 説明 | 使用例 |
|---|---|---|
| `trace_id` | リクエストを追跡する一意のID | `slog-otel` 等により自動付与 |
| `span_id` | 処理ブロックのID | `telemetry.StartSpan` により付与 |
| `action` | 実行されている具体的な処理名 | `telemetry.ActionImport`, `ProcessTranslation` |
| `resource_type` | 操作対象のリソースタイプ | `Task`, `User`, `Dictionary` |
| `resource_id` | 操作対象のID | `task-123`, `user-456` |
| `duration_ms` | 処理の所要時間 | `telemetry.StartSpan` の終了ログで自動付与 |
| `status` | 処理の結果状態 | `success`, `failure`, `pending` |
| `error_code` | アプリケーション固有のエラーコード | `NOT_FOUND`, `AUTH_FAILED` |
| `exception_class` | 例外の型名 | `SqliteError`, `LlmApiError` |
| `stack_trace` | スタックトレース | `telemetry.ErrorAttrs` で自動付与 |

## 2. ログ挿入の具体的手法

### 2.1. 外部 I/O (Database, API)
データベースアクセスや外部API呼び出しの前後で、以下のパターンを適用する。

```go
// 開始時
slog.DebugContext(ctx, "database query start",
    slog.String("query", sql),
    slog.Any("params", args),
)

// 終了時
if err != nil {
    slog.ErrorContext(ctx, "database query failed",
        telemetry.ErrorAttrs(err)...
    )
} else {
    slog.DebugContext(ctx, "database query completed",
        slog.Int("rows_affected", n),
    )
}
```

### 2.2. 条件分岐の「決定理由」
複雑なロジックにおいて、なぜその分岐に入ったかを記録する。

```go
if !isUserAdmin {
    slog.InfoContext(ctx, "access denied",
        slog.String("reason", "insufficient_role"),
        slog.String("user_id", userId),
    )
    return ErrForbidden
}
```

### 2.3. 状態変化 (State Change)
データのステータスが変わるタイミングで、変更前後の情報を記録する。

```go
slog.InfoContext(ctx, "task status changed",
    slog.String("task_id", taskId),
    slog.String("old_status", oldStatus),
    slog.String("new_status", newStatus),
)
```

### 2.4. パフォーマンス計測とエラーコンテキスト
`pkg/infrastructure/telemetry` のユーティリティを最大限活用する。

```go
func (s *service) Process(ctx context.Context) error {
    // パフォーマンス計測開始 (span.start, span.end ログが自動出力される)
    defer telemetry.StartSpan(ctx, telemetry.ActionProcess)()

    if err := s.doWork(ctx); err != nil {
        // エラー詳細（コード、型、スタックトレース）を含めて記録
        slog.ErrorContext(ctx, "process failed", telemetry.ErrorAttrs(err)...)
        return err
    }
    return nil
}
```

## 3. 実装上の留意点
- `slog.DebugContext` を多用し、通常時はノイズを抑えつつ、デバッグ時に詳細な情報を得られるようにする。
- 引数や戻り値に `context.Context` が含まれていることを確認し、必ず `Context` 経由でログを出力する。
- [`pkg/infrastructure/telemetry/span.go`](pkg/infrastructure/telemetry/span.go) で定義されている `ActionType` を適宜拡張する。
- 構造化ログとしての読みやすさを考慮し、メッセージは簡潔な名詞句（例: `span.start`, `db.query.success`）にする。
