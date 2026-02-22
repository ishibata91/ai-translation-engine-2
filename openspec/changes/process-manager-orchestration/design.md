# Design: Process Manager Orchestration

## Context
ProcessManagerの構想（`spec.md`）自体は存在していましたが、今回「JobQueueインフラ」が新設され、「各スライスはインフラを直接触らず、プロンプトの生成と結果の保存に特化する」という設計（Option 4: Slice-Owned Callbackパターン）が確定したため、ProcessManagerがそのハブとして実体化する必要があります。

## Goals / Non-Goals

**Goals:**
- 各スライスが定義する独自の Contract（`TermTranslatorInput`等）に合わせて、共有データ（Loaderの出力結果など）をマッピング・変換する責務を持つ。
- スライスの「Phase 1 (プロンプト生成)」を呼び出して `[]llm_client.Request` を受け取る。
- 生成したUUID（`ProcessID`）と共に `JobQueue.SubmitJobs` を呼び出し、LLM実行処理をインフラに委譲する。
- JobQueueから結果（`[]llm_client.Response`）を回収し、スライスの「Phase 2 (保存)」を呼び出す。
- UIからのWebSocket/SSEコネクションを管理し、インフラからの `ProgressNotifier` をブラウザへ中継する。

**Non-Goals:**
- ProcessManager自身が特定のドメインロジック（どう訳すか、どう要約するか）を持つこと。
- インフラ層のJobQueueの内部実装（SQLite操作や再試行処理）に干涉すること。

## Decisions

1. **メモリ上または一時SQLiteでの State 管理**:
   - ProcessManager は現在進行中の `ProcessID` と「それがどのスライスの、どの処理フェーズか」を管理する状態情報（State）を持ちます。これはメモリ上で持つか、または `process_state.db` のような軽量なKVSで管理します。今回はアプリ再起動時のBatch API復旧を考慮し、ProcessManager自身もSQLiteで状態をバックアップする構成（Active/Resumable State）とします。

2. **DTOのマッピング層としての ProcessManager**:
   - `LoaderSlice` の出力（`ExtractedData`）へ直接依存することを各スライスに禁じるため、ProcessManagerが仲介者として `Transformer` （変換器）の役割を果たします。これにより、完全な腐敗防止層(ACL)が実現します。

## Risks / Trade-offs

- **[Risk] ProcessManager（God Object）の肥大化**:
  マッピングコードや進行制御コードが集中し、ファイルが巨大化する。
  → **Mitigation**: `refactoring_strategy.md` に従い、ファイル分割ではなく同一ファイル内のプライベートメソッド層として処理を分割（SRP化）し、目次のように自己文書化することで人間の認知負荷を抑制する。
