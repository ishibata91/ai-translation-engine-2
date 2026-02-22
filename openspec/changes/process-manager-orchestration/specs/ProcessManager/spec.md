# ProcessManager Spec

## ADDED Requirements

### Requirement: スライスからJobQueueへのプロセス委譲 (Job Dispatch)
ProcessManagerは、連携対象のドメインスライスからプロンプトのリストを受け取り、一意のProcessIDと共にJobQueueへ非同期処理を委譲しなければならない。

#### Scenario: プロンプト生成後のジョブ投入
- **WHEN** `SummaryGenerator.PreparePrompts()` が10件のRequestを返した
- **THEN** ProcessManagerは新しいUUID（ProcessID）を発行し、そのRequestリストを `JobQueue.SubmitJobs` に渡す
- **AND** 自身のローカルな状態管理（メモリまたはDB）に、その ProcessID が「SummaryGeneratorの処理中」であることを記録する

### Requirement: JobQueueからスライスへの永続化コールバック (Result Persistence Callback)
ProcessManagerは、JobQueueで処理が完了したProcessIDを検知し、その結果リストを該当するドメインスライスの永続化関数へ引き渡さなければならない。

#### Scenario: ジョブ完了時のスライス保存
- **WHEN** JobQueue から対象の ProcessID の完了連絡を受けた
- **THEN** ProcessManager は `JobQueue.GetResults` を呼んで全 Response を取得する
- **AND** それを `SummaryGenerator.SaveResults` に渡してドメインDBへの保存を完了させる
- **AND** 保存成功後、`JobQueue.DeleteJobs` を呼んでインフラのキューからHard Deleteする

### Requirement: 異なるドメインスライス間のデータマッピング (Anti-Corruption Layer)
ProcessManagerは、前段のスライスからの出力を直接後段のスライスに渡さず、後段スライスの要求するContract（入力DTO）への変換を責務として負わなければならない。

#### Scenario: Loader出力からドメイン入力への変換
- **WHEN** `LoaderSlice` が `ExtractedData` を返した
- **THEN** ProcessManager はそれを `SummaryGeneratorInput` や `PersonaGenInput` などの専用DTOに詰め替える
- **AND** ドメインスライスは `ExtractedData` （`pkg/domain` レベルのグローバル参照）を一切インポートせずに動作する
