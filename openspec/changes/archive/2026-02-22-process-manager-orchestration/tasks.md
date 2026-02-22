# Tasks: Process Manager Orchestration

## 1. 進行状態 (State) 管理基盤の実装
- [x] `pkg/process_manager/store.go` を作成し、SQLite (`process_state.db`) を用いた進行中プロセスの永続化ロジックを実装する。
- [x] `ProcessManager` 構造体を定義し、起動時に `JobQueue` の `Recover` と自身の `State` を照合して処理を再開するリカバリロジックを実装する。

## 2. 外部インターフェース (UI連携) の整備
- [x] WebSocket ハンドラを作成し、特定の `CorrelationID` (TraceID相当) に基づく進捗イベントのブロードキャスト機能を実装する。
- [x] UI からの「プロセス開始」リクエストを受け取り、`ProcessID` を発行してレスポンスする API を作成する。

## 3. オーケストレーションロジックの実装 (VSA Compliance)
- [x] スライスの「Phase 1 (プロンプト生成)」を呼び出し、`JobQueue.SubmitJobs` へ登録する一連の処理（Executor）を実装。
- [x] `ProgressNotifier` インターフェースを実装（Hub化）し、JobQueue からの進捗を UI (WebSocket) へ転送するブリッジを作成。
- [x] ジョブ完了を検知（ポーリングまたは同期待ちの終了）し、結果をスライスの「Phase 2 (保存)」へ引き渡すコールバックフローの実装。

## 4. DTO マッピング (Anti-Corruption Layer)
- [x] 各スライスの `InputDTO` を生成するための Mapper 関数群を `ProcessManager` 内に実装する。
- [x] `LoaderSlice` の出力から `TermTranslator`, `PersonaGen` 等の各ドメイン用コンテキストへの変換を定義。

## 5. 結合テスト
- [x] テスト用 SQLite を用い、アプリ再起動をシミュレートした状態でのジョブ・レジューム成功（バッチポーリングの再開と保存完了）を検証する。
