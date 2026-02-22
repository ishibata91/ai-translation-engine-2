# Tasks: Process Manager Orchestration

## 1. ProcessManager のハブ機能作成
- [ ] `pkg/process_manager` のメインコントローラを作成し、メモリ上または軽量SQLite（`process_state.db`）で `ProcessID` と進行中のフェーズ（どのスライスの処理か）を管理するState構造体を定義する。
- [ ] UI（WebSocket/SSE用ハンドラ等）からの「翻訳開始」イベントの受け口を整備する。

## 2. DTO変換（Anti-Corruption Layer）の実装
- [ ] `LoaderSlice` の出力（`ExtractedData`）を受け取り、後続スライスが必要とする `SummaryGeneratorInput`, `PersonaGenInput`, `TermTranslatorInput` にマッピングする Mapper 関数群を実装する。

## 3. Executor / Runner ループの実装
- [ ] スライスの「Phase 1」関数を呼び出して生成した `Request` を `JobQueue.SubmitJobs` でインフラに投げるロジックを実装。
- [ ] JobQueue側が実装する `ProgressNotifier` をラップし、ProcessManagerを通じてUIのWebSocketまで進捗率を流し込むパイプラインを実装する。
- [ ] JobQueueが完了ステータスになったプロセスIDについて、結果を取得してスライスの「Phase 2」関数をコールし、その後 `JobQueue.DeleteJobs` でHard Deleteする完了ライフサイクルを実装する。

## 4. 結合とテスト
- [ ] 模擬のJobQueueと模擬のドメインスライスを用いて、ProcessManagerがデータ変換・ジョブ投入・回収・結果保存の一連の流れをオーケストレートできるかテストする。
