# Design: Generic Job Queue Infrastructure

## Context
「24時間かかるバッチAPIのポーリング」という極端な非同期性とアプリケーションの再起動耐性を両立しつつ、VSAの「各スライスの永続化はスライス内で完結する」大原則を守るため、**インフラ側のキューにはドメイン知識を一切持たせない**（どのスライスからのリクエストかを知らない）設計が不可欠です。

**前提 Change との依存関係**:
この Change は `llm-client-bulk-sync` Change の完了を前提とします。特に、同期バルク処理ヘルパー `ExecuteBulkSync` および ConfigStore による `llm.bulk_strategy`（`"batch"` / `"sync"`）制御の実装が先行して完了している必要があります。

## Goals / Non-Goals

**Goals:**
- `ProcessID` (ただのUUID) と `Request` (LLMプロンプト) だけを保持するSQLiteベースのインフラキューを作成する。
- ConfigStore の `llm.bulk_strategy` キーを参照し、`"sync"` なら `ExecuteBulkSync`、`"batch"` ならプロバイダの `SubmitBatch` ＆ ポーリング をワーカーがバックグラウンドで実行する。ローカルLLM使用時は `"batch"` を選んでも `"sync"` へフォールバックする。
- 完了した結果は ProcessManager を経由して各スライスに渡され、スライスが自身のDB（`summaries`等）に保存した後に、インフラのキューからは即座に削除（Hard Delete）することで、インデックスの肥大化を防ぎ検索パフォーマンスを永続的に保つ。
- ワーカーの処理進捗を `ProgressNotifier` 経由で UI (ProcessManager) へ透過的にフィードバックする。

**Non-Goals:**
- キューワーカーが `TermTranslator` や `SummaryGenerator` などの具体的な結果の解釈や、ドメインDBへの書き込みを行うこと。
- スライス間でデータを共有すること。

## Decisions

1. **インフラ層の独立 SQLite データベース (`llm_jobs.db`)**:
   - `id` (UUID), `process_id` (UUID), `request_json`, `status` (PENDING, IN_PROGRESS, COMPLETED, FAILED), `batch_job_id` (バッチAPI用), `response_json` 等をカラムに持つ単一テーブルを作成。
   - `process_id` にインデックスを張り、ProcessManager が `GetResults(ProcessID)` で高速に結果を刈り取れるようにする。

2. **Hard Delete モデリング (消費即削除)**:
   - レコードは「保存・処理待ち」の一時的なものとしてのみ扱う。
   - `GetResults` され、スライス側での保存が成功したジョブは、即座に JobQueue データベースから物理削除(`DELETE`)する。
   - これにより、システムを長期間運用しても `llm_jobs.db` は常にクリーンで超高速な状態を維持する。

3. **透過的なUIプログレス通知**:
   - ワーカープロセス起動時に `ProgressNotifier` インターフェースを受け取る。
   - `llm.bulk_strategy` が `"sync"` の場合（`ExecuteBulkSync` 使用時）は1件処理ごとに `OnProgress` を発火する。
   - `llm.bulk_strategy` が `"batch"` の場合は `GetBatchStatus` のポーリング結果をそのまま `OnProgress` に流す。
   - ローカルLLM使用時は `"sync"` フォールバックが適用されるため、常に前者のプログレス方式を採用する。

## Risks / Trade-offs

- **[Risk] SQLiteデータベースのロック/競合**:
  バックグラウンドワーカーとProcessManager（登録処理）が同時にアクセスする。
  → **Mitigation**: `PRAGMA journal_mode=WAL` と十分な `busy_timeout` を設定し、インフラ層専用のSQLiteとして分離することで安全性を担保する。またキューサイズが極小に保たれる（Hard Delete）ため問題になりにくい。
