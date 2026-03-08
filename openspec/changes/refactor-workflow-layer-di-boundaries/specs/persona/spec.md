## MODIFIED Requirements

### Requirement: ペルソナ生成ジョブ提案（Phase 1）はタスク境界で実行されなければならない
ペルソナ生成の Phase 1（`PreparePrompts`）は、UI 同期呼び出しではなく controller から workflow を経由したタスク境界で実行されなければならない。システムは開始から完了/失敗まで同一 task ID で追跡可能でなければならず、処理中のログは `specs/architecture.md` に準拠した構造化ログとして出力しなければならない。persona slice は workflow から契約経由で呼び出され、runtime 制御や UI 都合の状態遷移を直接持ってはならない。

#### Scenario: UI 起点で Phase 1 が workflow 配下の task として実行される
- **WHEN** `MasterPersona` からペルソナ生成開始要求が送信される
- **THEN** controller は workflow を介して `PreparePrompts` を非同期実行しなければならない
- **AND** task 状態を `Pending` -> `Running` -> `Completed` または `Failed` に遷移させなければならない

#### Scenario: persona slice は queue を直接制御しない
- **WHEN** `PreparePrompts` または `SaveResults` が実行される
- **THEN** persona slice は queue への enqueue、resume、cleanup を実行してはならない
- **AND** 当該制御は runtime 契約を通じて workflow が担わなければならない

#### Scenario: LLM 統合前段でも同一 task ID で追跡できる
- **WHEN** Phase 1 のリクエスト生成処理が完了する
- **THEN** システムは将来の `pkg/llm` 連携に引き継げる単一 task ID を保持しなければならない
- **AND** 完了時点でリクエスト件数サマリを `info` ログに記録しなければならない

### Requirement: 独立性: ペルソナ生成データの受け取りと独自DTO定義
本 slice は、独立性を確保する Anti-Corruption Layer パターンに従い、自前の入力 DTO と保存 DTO を契約として公開しなければならない。workflow は parser や runtime から受け取ったデータを当該 DTO へ変換し、persona slice は外部 DTO や外部 runtime 固有型へ直接依存してはならない。

#### Scenario: workflow が persona 用 DTO を組み立てる
- **WHEN** workflow が parser の出力を受け取る
- **THEN** workflow は persona slice 専用の入力 DTO を組み立てて渡さなければならない
- **AND** persona slice は parser の DTO を直接参照してはならない

#### Scenario: runtime 結果は外部 DTO のまま保存契約へ渡されない
- **WHEN** workflow が runtime から LLM 実行結果を取得する
- **THEN** workflow は persona slice の保存契約に適合するデータへ変換してから渡さなければならない
- **AND** persona slice は runtime の永続化 DTO や request state を理解してはならない
