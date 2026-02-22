# Proposal: Progress Notifier Infrastructure

## Why

現在、`job-queue-infrastructure` の設計において「ワーカープロセスが処理進捗を UI へフィードバックする」ために `ProgressNotifier` インターフェースが必要とされている。

しかし、このインターフェースを JobQueue パッケージ内に閉じ込めると、将来的に他のスライスやインフラコンポーネントが同様の進捗通知を行いたい場合に **JobQueue への不正な依存** が発生する。これは VSA の依存方向に違反し、スライス間の密結合を引き起こす。

`refactoring_strategy.md §5` の WET vs Shared Kernel 判断基準に従うと、進捗通知は「ドメイン知識を持たない **技術的関心事**（Infrastructure/Utility レベル）」に該当するため、`Shared/Core レイヤーへDRYに集約`することが正しい設計判断となる。

## What Changes

### New Capabilities

- `progress-notifier`: `pkg/infrastructure/progress` パッケージとして汎用進捗通知インターフェースと ProgressEvent 型を定義する。任意のスライス・インフラコンポーネントがこのインターフェースを DI で受け取ることで、実装（WebSocket / SSE / ログ出力 等）に依存せず進捗フィードバックを実現できる。

### Modified Capabilities

（なし）

## Impact

- **影響コンポーネント**:
  - `pkg/infrastructure/progress/` （新規パッケージ）
  - `openspec/changes/job-queue-infrastructure/design.md` （Decision 3/5 を本 Change への参照に更新）
  - 将来的に進捗通知を必要とする全スライス（`SummaryGeneratorSlice`, `TermTranslatorSlice` 等）が DI 経由で利用
- **依存ライブラリ**: 追加なし（Go 標準ライブラリのみ）
- **破壊的変更**: なし（新規パッケージの追加のみ）
