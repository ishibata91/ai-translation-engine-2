# Tasks: Generic Job Queue Infrastructure

> **前提**: `llm-client-bulk-sync` Change（`ExecuteBulkSync` および `llm.bulk_strategy` ConfigStore 実装）が完了してから着手すること。

## 1. 共通 Queue Infrastructure の実装
- [ ] `pkg/infrastructure/job_queue` ディレクトリを作成する。
- [ ] `queue.go` を作成し、SQLite接続管理、`llm_jobs` テーブルの初期化処理（DDL・PRAGMA設定等）を実装する。
- [ ] `ProcessID` (UUID) ベースの `SubmitJobs(ctx, processID, reqs)` と `GetResults(ctx, processID)` を実装する。
- [ ] 完了済み結果を取得した際に物理削除する Hard Delete ロジック (`DeleteJobs(ctx, processID)`) を実装する。

## 2. ワーカーとLLMClientの統合
- [ ] `worker.go` を作成し、起動時に ConfigStore から `llm.bulk_strategy` を読み込んで実行モードを決定するロジックを実装する。
- [ ] `"sync"` 戦略時: `ExecuteBulkSync`（`llm-client-bulk-sync` で実装済）を使ってバックグラウンド処理を実行する。
- [ ] `"batch"` 戦略時: xAI等の Batch APIにおける `BatchJobID` のポーリングロジックをワーカー内に実装する。
- [ ] ローカルLLMプロバイダの場合、`"batch"` 設定であっても `"sync"` に強制フォールバックするロジックを実装する。
- [ ] `ProgressNotifier` をワーカーに注入し、`"sync"` 時は1件ごと / `"batch"` 時はポーリング結果から、UI 用のプログレス更新を発火する仕組みを実装する。

## 3. テストの作成
- [ ] インメモリSQLite (`:memory:`) を用いた `job_queue_test.go` を作成する。
- [ ] `SubmitJobs` → (ワーカー処理) → `GetResults` → `DeleteJobs` の包括的なライフサイクルテスト（Hard Deleteされテーブルが空になること）を実行する。
- [ ] バッチAPIモードにおいて、一時的なネットワーク切断やリトライなどに対する耐障害性テストを記述する。
