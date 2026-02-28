# Tasks: frontend-job-manager

## 1. データベース・永続化レイヤー (Go)
- [x] 1.1 `schema.sql` 等にタスクを保持する `frontend_tasks` テーブルと、状態復元用メタデータを保存する `task_metadata` (もしくはJSON型のカラム) を定義する。
- [x] 1.2 DBの永続化モデルに対応する構造体 `models.Task` と `models.TaskMetadata` を新たに作成する。
- [x] 1.3 新規タスクを作成する関数 `InsertTask(ctx, task)` を実装する。
- [x] 1.4 タスクの状態（ステータス、フェーズ、進捗、エラー）を更新する関数 `UpdateTask(ctx, taskID, status, phase, progress, errorMsg)` を実装する。
- [x] 1.5 中断時からの再開用パラメータを保存・取得するメタデータ操作関数 `SaveMetadata`/`GetMetadata` を実装する。
- [x] 1.6 初期ロード・一覧表示用に `GetAllTasks` と `GetActiveTasks` のクエリを実装する。

## 2. タスクマネージャーのコア実装 (Go)
- [x] 2.1 新たに `TaskManager`（または `FrontendJobManager`）構造体とパッケージを設計し、メモリ上のアクティブタスク一覧（スレッドセーフなマップ等）とDBリポジトリを管理させる。
- [x] 2.2 `TaskManager` の初期化処理（`NewManager`）を実装し、アプリ起動時にDBから未完了タスクを読み込んでメモリにマウントする。
- [x] 2.3 アプリの異常終了などでステータスが `Running` のまま残っているタスクを、初期化時に `Paused` または `Error` へ自動補正するロジックを実装する。
- [x] 2.4 新規タスクを登録・開始する `AddTask` メソッドを新設する。DBへのレコード登録と初期フェーズの設定後、処理をゴルーチンでバックグラウンド実行すること。
- [x] 2.5 タスクの進捗（`Progress` %等）変化時は、Wails の `EventsEmit` は高頻度（例：ループ1回ごと）に即時送信し、一方でSQLite等DBへの保存はスロットリング・デバウンス（例：数秒おき、または一定割合進行時のみ）機能を用いて分離させ、負荷を抑える。
- [x] 2.6 タスク完了時(`Completed`)、失敗時(`Failed`)、キャンセル時(`Cancelled`) に、最終ステータスを確実にDBへ更新し、後処理を行うロジックを実装する。

## 3. フェーズ完了時のデータ一括保存と通知機能 (Go/Wails)
- [x] 3.1 タスクの全体進捗を即座に通知する `task:updated` イベントの発火処理を `TaskManager` に実装する。（DB保存とは同期させず、高頻度なemitによってUIへリアルタイム反映させる）
- [x] 3.2 複数フェーズを跨ぐタスク向けに、あるフェーズが終わったタイミング（例：ペルソナ抽出処理が全て終わった後）でのみデータをDBへ保存し、`EmitPhaseCompleted(taskID, phaseName, dataSummary)` 等でフロントへ通知するインタフェースを整備する。（※ループ中の1件ごとのDB保存はN+1問題回避のため禁止）
- [x] 3.3 （連携先タスク実装での遵守事項）大きなループ処理において、1件完了するたびに `Progress` の即時通知(EventsEmit)を行い、DBへの「単体データの保存」は行わない（進捗のDB保存は TaskManager 側のスロットリング機能に任せる）よう処理を実装する。（※本体データの書き込みはフェーズ完了時に一括で行うこと）

## 4. レジューム（再開）フローと Wails Bindings (Go)
- [x] 4.1 React側から呼び出し可能な Wails バインディングとして、現在状態を取得する `GetActiveTasks()` をエクスポート（Wailsのバインドに登録）する。
- [x] 4.2 中断されたタスクを再開するための `ResumeTask(taskID string)` バインディングを追加する。
- [x] 4.3 `ResumeTask` 呼び出し時、DB から現在の `Phase` と `Metadata` を復元し、適切な位置から処理を再開するディスパッチャを実装する。
- [x] 4.4 ジョブの中止処理として Wails バインディング `CancelTask(taskID)` と、`context.CancelFunc` を用いたゴルーチン停止連携を実装する。

## 5. フロントエンド型定義とタスク管理の基盤 (React)
- [x] 5.1 `frontend/types` ディレクトリに、Goバックエンドに合わせて `Phase` や `Metadata` を包含する `FrontendTask` 等の型定義を新規作成する。
- [x] 5.2 既存の `frontend/src/store/uiStore.ts` を拡張するか、同ディレクトリに `taskStore.ts` を作成し、アプリ上のタスク状態（リスト・進捗・エラー等）を管理する基盤を作る。
- [x] 5.3 アプリ初期化時（Dashboardマウント時等）、`wailsjs/go/.../GetActiveTasks()` を呼び出して、保存されている初期状態のタスク一覧をストアへ取り込む。

## 6. イベント購読と画面間ルーティング・UI連携 (React)
- [x] 6.1 `wailsjs/runtime/EventsOn("task:updated")` と `"task:phase_completed"` の購読処理を新設する。UIは進捗に応じたバーの表示を行い、フェーズ完了通知が届いた場合は関連する一覧データの再フェッチ（一括更新）を行う。
- [x] 6.2 ダッシュボードのタスク一覧などからアクティブなタスク行をクリックした際に、タスクの `Phase` から適切な画面のURL（と開くべきタブ）を特定して遷移するロジック（react-router連携）を構築する。
- [x] 6.3 アプリケーション内の各詳細ビュー等に、未完了(`Paused`)のタスクを再開させる `Resume` ボタン（や関連UI）を新たに設計・配置し、`ResumeTask` APIと接続する。
- [x] 6.4 `Paused` または `Error` になっているタスクのUI表示を区別（色や警告アイコンなど）し、実行中の `Running` 状態と視覚的に明確な差異を持たせる表現を実装する。
