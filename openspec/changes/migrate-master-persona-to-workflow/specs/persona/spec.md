## MODIFIED Requirements

### Requirement: ペルソナ生成ジョブ提案（Phase 1）はタスク境界で実行されなければならない
ペルソナ生成の Phase 1（`PreparePrompts`）は、UI 同期呼び出しではなく controller から workflow を経由したタスク境界で実行されなければならない。persona slice は request 準備と保存に集中し、queue 制御や task 状態遷移を持ってはならない。

#### Scenario: UI 起点で Phase 1 が workflow 配下の task として実行される
- **WHEN** `MasterPersona` からペルソナ生成開始要求が送信される
- **THEN** controller は workflow を介して `PreparePrompts` を非同期実行しなければならない

#### Scenario: persona slice は queue を直接制御しない
- **WHEN** `PreparePrompts` または `SaveResults` が実行される
- **THEN** persona slice は queue への enqueue、resume、cleanup を実行してはならない
