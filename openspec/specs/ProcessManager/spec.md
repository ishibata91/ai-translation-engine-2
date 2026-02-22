# ProcessManager Spec

## Purpose
ProcessManagerは、個別のVertical Slice（ドメインロジック）と共有インフラストラクチャ（JobQueue, LLMClient等）の間のオーケストレーションを担う。データのマッピング、状態の管理、および再起動後の処理継続を責務とする。

## Requirements

### Requirement: スライスからJobQueueへのプロセス委譲 (Job Dispatch)
ProcessManagerは、連携対象のドメインスライスからプロンプトのリストを受け取り、一意の ProcessIDと共にJobQueueへ非同期処理を委譲しなければならない (MUST)。

#### Scenario: プロンプト生成後のジョブ投入
- **WHEN** `SummaryGenerator.PreparePrompts()` が10件のRequestを返した
- **THEN** ProcessManagerは新しいUUID（ProcessID）を発行し、そのRequestリストを `JobQueue.SubmitJobs` に渡す
- **AND** 自身のローカルな状態管理（`process_state.db`）に、その ProcessID が「SummaryGeneratorの処理中」であることを記録する

### Requirement: JobQueueからスライスへの永続化コールバック (Result Persistence Callback)
ProcessManagerは、JobQueueで処理が完了したProcessIDを検知し、その結果リストを該当するドメインスライスの永続化関数へ引き渡さなければならない (MUST)。

#### Scenario: ジョブ完了時のスライス保存
- **WHEN** JobQueue から対象の ProcessID の完了連絡を受けた
- **THEN** ProcessManager は `JobQueue.GetResults` を呼んで全 Response を取得する
- **AND** それを `SummaryGenerator.SaveResults` に渡してドメインDBへの保存を完了させる
- **AND** 保存成功後、`JobQueue.DeleteJobs` を呼んでインフラのキューからHard Deleteする

### Requirement: 異なるドメインスライス間のデータマッピング (Anti-Corruption Layer)
ProcessManagerは、前段のスライスからの出力を直接後段のスライスに渡さず、後段スライスの要求するContract（入力DTO）への変換を責務として負わなければならない (MUST)。

#### Scenario: Loader出力からドメイン入力への変換
- **WHEN** `LoaderSlice` が `ExtractedData` を返した
- **THEN** ProcessManager はそれを `SummaryGeneratorInput` や `PersonaGenInput` などの専用DTOに詰め替える
- **AND** ドメインスライスは `pkg/domain` レベルのグローバル参照を一切受容せず、独自の InputDTO のみを使用して動作する

### Requirement: UIへのリアルタイム通知と進捗中継 (UI Notification Relay)
ProcessManagerは、JobQueue や各スライスから発生する進捗イベント（`ProgressEvent`）を収集し、WebSocket / SSE 経由で UI (React) へ即座に配信しなければならない (MUST)。

#### Scenario: 進捗状況のウェブ同期
- **WHEN** JobQueue のワーカーが翻訳進捗を `notifier.OnProgress` で発火した
- **THEN** ProcessManager はそのイベントをキャッチし、接続中のブラウザへ JSON 形式で転送する

### Requirement: プロセス中断後の自動レジューム (Process Resume)
ProcessManagerは、アプリケーションの再起動や予期せぬ終了後、DBに保存された進行状態（State）を読み込み、未完了のジョブのステータス確認から処理を再開しなければならない (MUST)。

#### Scenario: 再起動後のバッチリカバリ
- **WHEN** 再起動時に `process_state.db` に `IN_PROGRESS` の `BatchJobID` が残っていた
- **THEN** ProcessManager は `InputFile` 情報を維持したまま、JobQueue を介さず（または JobQueue の Recover を前提に）該当ジョブのポーリングを再開し、完了後のコールバックフローに合流する
