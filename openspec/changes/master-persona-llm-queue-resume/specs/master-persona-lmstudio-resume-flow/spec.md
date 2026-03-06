## ADDED Requirements

### Requirement: MasterPersona は LM Studio 専用の段階実行フローを提供しなければならない
システムは MasterPersona 実行において、`request enqueue` -> `lmstudio dispatch` -> `persona save` の 3 段階を単一 task ID で実行・追跡しなければならない。プロバイダは `lmstudio` 以外を受け付けてはならない。

#### Scenario: LM Studio 以外は開始拒否される
- **WHEN** MasterPersona 開始時の provider が `lmstudio` 以外である
- **THEN** システムはタスクを開始せず、無効な provider エラーを返さなければならない

#### Scenario: 段階実行が単一 task ID で追跡される
- **WHEN** MasterPersona を開始する
- **THEN** システムは enqueue/disptach/save の各段階を同一 task ID に紐づけて状態更新しなければならない
