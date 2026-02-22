# Proposal: Generic Job Queue Infrastructure

## Problem & Motivation
現在のアーキテクチャでは、設定によって「数秒で終わる同期API」と「最大24時間かかるバッチAPI（xAI/OpenAI）」が切り替わる可能性があります。
しかし、各ドメインスライス（SummaryGenerator等）が直接 `LLMClient.SubmitBatch` を呼び出し、そのステータスをメモリ上でポーリングして待つような設計にしてしまうと、アプリの再起動やクラッシュ時に `BatchJobID` が失われ、処理の再開が不可能になります。
さらに、バッチAPIのエラーリトライや進捗状況のUI通知といった「インフラ的な泥臭い処理」がすべてのドメインスライスに散らばり、Vertical Slice Architecture (VSA) の純粋性を著しく損ないます。

## Proposal
インフラ層（`pkg/infrastructure/job_queue`）に、各スライスのドメイン知識を一切持たない汎用的なSQLiteベースのジョブキューを新設します。
各スライスは「プロセスID」と「LLMリクエスト」のペアを生成するだけであり、ProcessManager がこのキューに登録し、実行させます。キューのワーカーはバックグラウンドでSync/Batchを問わずLLMと通信し、完了したジョブの結果を受け取ったProcessManagerが最終的にスライスに再保存を依頼する「Slice-Owned Callback パターン（インフラ委譲）」を実現します。

## Capabilities

### New Capabilities
- `JobQueue`: アプリケーション再起動時にも回復可能なインフラ層の永続化ジョブキューワーカー。プロセスIDベースで完全に非結合な状態管理と、UIへの透過的な進捗通知（`ProgressNotifier`）を提供する。

## Impact
- `pkg/infrastructure/job_queue` パッケージの新設と専用SQLite（`llm_jobs.db`等）の作成。
- `ProcessManager` にインフラ層のJobQueueとの調整責務が付与される。
- UI側でのリアルタイムなバッチ進捗表示の基盤が完成する。
