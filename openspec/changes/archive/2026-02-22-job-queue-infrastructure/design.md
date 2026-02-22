# Design: Generic Job Queue Infrastructure

## Context

「24時間かかるバッチAPIのポーリング」という極端な非同期性とアプリケーションの再起動耐性を両立しつつ、VSAの「各スライスの永続化はスライス内で完結する」大原則を守るため、**インフラ側のキューにはドメイン知識を一切持たせない**（どのスライスからのリクエストかを知らない）設計が不可欠です。

**前提 Change との依存関係**:
この Change は `llm-client-bulk-sync` Change の完了を前提とします。特に、同期バルク処理ヘルパー `ExecuteBulkSync` および ConfigStore による `llm.bulk_strategy`（`"batch"` / `"sync"`）制御の実装が先行して完了している必要があります。

## Goals / Non-Goals

**Goals:**
- `process_id` (ただのUUID) と `request_json` (LLMプロンプトJSON) だけを保持するSQLiteベースのインフラキュー (`llm_jobs.db`) を作成する。
- ConfigStore の `llm.bulk_strategy` キーを参照し、`"sync"` なら `ExecuteBulkSync`、`"batch"` ならプロバイダの `SubmitBatch` ＆ ポーリング をワーカーがバックグラウンドで実行する。ローカルLLM使用時は `"batch"` を選んでも `"sync"` へフォールバックする。
- 完了した結果は ProcessManager を経由して各スライスに渡され、スライスが自身のDB（`summaries`等）に保存した後に、インフラのキューからは即座に削除（Hard Delete）することで、インデックスの肥大化を防ぎ検索パフォーマンスを永続的に保つ。
- ワーカーの処理進捗を `ProgressNotifier` 経由で UI (ProcessManager) へ透過的にフィードバックする。

**Non-Goals:**
- キューワーカーが `TermTranslator` や `SummaryGenerator` などの具体的な結果の解釈や、ドメインDBへの書き込みを行うこと。
- スライス間でデータを共有すること。

## Decisions

### 1. インフラ層の独立 SQLite データベース (`llm_jobs.db`)

`database_erd.md` の Job Queue Infrastructure セクションに定義されたスキーマを使用する（テーブル名・カラム名・型はERDを正として扱う）。

| カラム名        | 型       | 説明                                                      |
| :-------------- | :------- | :-------------------------------------------------------- |
| `id`            | TEXT     | ジョブID (UUID), PRIMARY KEY                              |
| `process_id`    | TEXT     | 処理単位ID (UUID), INDEX。ProcessManager が刈り取りに使用 |
| `request_json`  | TEXT     | LLMリクエスト (JSON)                                      |
| `status`        | TEXT     | `PENDING` / `IN_PROGRESS` / `COMPLETED` / `FAILED`        |
| `batch_job_id`  | TEXT     | Batch API ジョブID (batch戦略時のみ使用, nullable)        |
| `response_json` | TEXT     | 完了時のLLMレスポンス (JSON, nullable)                    |
| `error_message` | TEXT     | エラーメッセージ (FAILED時のみ, nullable)                 |
| `created_at`    | DATETIME | 登録日時                                                  |
| `updated_at`    | DATETIME | 最終更新日時                                              |

- `process_id` にインデックスを張り、ProcessManager が `GetResults(ProcessID)` で高速に結果を刈り取れるようにする。
- `schema_version` テーブルを同居させてマイグレーション管理を行う（`config.db` と同様のパターン）。

### 2. Hard Delete モデリング (消費即削除)

VSA原則の「インフラ層の独立した永続化」に従い、キューは一時的な保管庫として機能する。

- レコードは「保存・処理待ち」の一時的なものとしてのみ扱う。
- `GetResults` され、スライス側での保存が成功したジョブは、即座に JobQueue データベースから物理削除（`DELETE FROM llm_jobs WHERE id = ?`）する。
- Soft Delete（論理削除フラグや`deleted_at`カラム）は**一切使用しない**。
- これにより、システムを長期間運用しても `llm_jobs.db` は常にクリーンで超高速な状態を維持する。

### 3. 透過的なUIプログレス通知

- ワーカープロセス起動時に `ProgressNotifier` インターフェースを受け取る（DI経由）。
- `llm.bulk_strategy` が `"sync"` の場合（`ExecuteBulkSync` 使用時）は1件処理ごとに `OnProgress` を発火する。
- `llm.bulk_strategy` が `"batch"` の場合は `GetBatchStatus` のポーリング結果をそのまま `OnProgress` に流す。
- ローカルLLM使用時は `"sync"` フォールバックが適用されるため、常に前者のプログレス方式を採用する。

### 4. 構造化ログ・Context伝播 (refactoring_strategy.md §6.2 / §7 準拠)

- **全ての公開メソッドは第一引数に `ctx context.Context` を受け取り**、内部処理にも伝播させる。
- `slog.DebugContext(ctx, ...)` / `slog.InfoContext(ctx, ...)` 等の `Context` 付きメソッドを必ず使用し、OpenTelemetry の TraceID/SpanID が自動付与されるようにする。
- ワーカー関数および `GetResults`・`Enqueue` 等の主要メソッドに Entry/Exit ログを実装する（`"ENTER MethodName"` / `"EXIT MethodName"` + 引数・戻り値）。
- ログには `"slice": "JobQueue"` の属性を付与して横断フィルタリングを可能にする。

```json
{"level":"DEBUG","msg":"ENTER Enqueue","trace_id":"...","slice":"JobQueue","args":{"process_id":"...","job_count":3}}
{"level":"DEBUG","msg":"EXIT Enqueue","trace_id":"...","slice":"JobQueue","result":{"inserted":3,"elapsed":"2ms"}}
```

### 5. 外部パッケージの利用 (progress-notifier)

進捗通知には `pkg/infrastructure/progress` パッケージの `ProgressNotifier` インターフェースを DI で受け取って使用する。

- **CorrelationID の指定**: `ProgressEvent` の `CorrelationID` には、ProcessManager から渡された `ProcessID` (UUID) を設定する。これにより UI 側で適切にグループ化される。
- **通知タイミング**:
  - `llm.bulk_strategy = "sync"`：1件の処理完了ごとに `StatusInProgress` を通知。
  - `llm.bulk_strategy = "batch"`：ポーリング結果（件数またはパーセンテージ）を `ProgressEvent` に変換して通知。

```go
// 利用イメージ
func (w *worker) run(ctx context.Context) {
    // ...
    w.notifier.OnProgress(ctx, progress.ProgressEvent{
        CorrelationID: w.processID,
        Status:        progress.StatusInProgress,
        Message:       "Processing...",
    })
}
```

## Risks / Trade-offs

- **[Risk] SQLiteデータベースのロック/競合**:
  バックグラウンドワーカーとProcessManager（登録処理）が同時にアクセスする。
  → **Mitigation**: `PRAGMA journal_mode=WAL` と十分な `busy_timeout`（最低5秒）を設定し、インフラ層専用のSQLiteとして分離することで安全性を担保する。またキューサイズが極小に保たれる（Hard Delete）ため問題になりにくい。

- **[Risk] 再起動時の `IN_PROGRESS` ジョブ**:
  アプリ再起動時に `status = 'IN_PROGRESS'` のまま残ったジョブが再実行されない。
  → **Mitigation**: 起動時に `status = 'IN_PROGRESS'` を `status = 'PENDING'` にリセットするリカバリー処理をWorker起動ロジックに含める。

- **[Trade-off] キューの一時保管責務とVSAの自律性**:
  このキューはインフラ層（技術的関心事）として `Shared/Core レイヤーへDRYに集約` する（refactoring_strategy.md §5 WET vs Shared Kernel の判断基準に従う）。各ドメインスライスはキューを直接操作せず、ProcessManagerを介してのみ利用する。
