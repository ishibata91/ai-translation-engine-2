# JobQueue Spec

## ADDED Requirements

### Requirement: プロセスIDに基づくジョブの登録と完全な非結合実行
インフラストラクチャは、呼び出し元スライスのドメイン知識や保存先のテーブル（スキーマ）を一切知らずに、「任意のUUID (ProcessID)」と「LLMリクエストの配列」のみを受け取り、永続化キューに保存して実行を担保しなければならない。

#### Scenario: 汎用的なジョブ登録
- **WHEN** ProcessManagerが任意の `ProcessID` と共に LLMリクエストを `SubmitJobs` した
- **THEN** インフラの SQLite (`llm_jobs` テーブル) に状態 `PENDING` として安全に永続化される
- **AND** スライスのテーブルスキーマ（`summaries`等）には一切依存しない

### Requirement: 透過的なUIプログレス更新
ワーカーはSync/Batch実行モードに関わらず、受け取った `ProgressNotifier` を通じて定期的にUI（ProcessManager経由）に進捗状況(`OnProgress`)をプッシュしなければならない。

#### Scenario: バッチAPIのリアルタイム進捗通知
- **WHEN** Batch APIを用いたポーリングがX分おきに発生する
- **THEN** ポーリングで得た `status.Progress` 率がそのまま `notifier.OnProgress` に渡る
- **AND** 呼び出し元のスライス（ドメインロジック）は通知処理に一切関与しない

### Requirement: 結果回収時の Hard Delete (消費即削除)
結果の完了確認およびUI層からの回収が終わったプロセスIDのジョブデータは、インデックス肥大化を防ぐためにデータベースから即座に物理削除されなければならない。

#### Scenario: 健全なインデックスの維持
- **WHEN** ProcessManager が `GetResults` でLLM処理結果を受け取り、スライスのDBに保存成功した後に `DeleteJobs(ProcessID)` を呼び出す
- **THEN** 該当のジョブとレスポンスデータが `llm_jobs` テーブルから完全に DELETE される
- **AND** ストレージと検索パフォーマンスが常に O(1) に近い状態に保たれる
