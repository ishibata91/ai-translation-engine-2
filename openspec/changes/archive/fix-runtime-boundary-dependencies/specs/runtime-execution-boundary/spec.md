## ADDED Requirements

### Requirement: runtimeは外部I/O実行と実行制御に専念しなければならない
システムは、`pkg/runtime/**` を外部 I/O 実行と実行制御基盤に専念させなければならない。runtime は `workflow`、`slice`、`artifact` を直接 import してはならず、ユースケース進行決定や slice 固有判断を持ってはならない。

#### Scenario: runtimeが外部I/Oを実行する
- **WHEN** 開発者が LLM、外部 API、ファイル、secret、config を使う実処理を実装する
- **THEN** 当該実処理は runtime が担当しなければならない
- **AND** runtime は workflow や slice の具象実装へ直接依存してはならない

### Requirement: runtimeは中立な実行要求と結果DTOでworkflowと連携しなければならない
システムは、runtime と workflow の連携を、特定 slice 固有 DTO を含まない中立な実行要求および結果 DTO で構成しなければならない。

#### Scenario: workflowがruntimeへ実行要求を渡す
- **WHEN** workflow が外部実行を runtime へ依頼する
- **THEN** runtime は中立的な契約 DTO を受け取って処理しなければならない
- **AND** runtime は slice 固有の保存判定や UI 状態解釈を行ってはならない

### Requirement: runtime配下のテストも同じ責務境界に従わなければならない
システムは、`pkg/runtime/**` 配下の test code においても runtime の責務境界を維持しなければならない。workflow や slice への直接依存をテスト都合で許してはならない。

#### Scenario: runtime配下のテストで境界違反を追加する
- **WHEN** 開発者が `pkg/runtime/**` 配下の test code で workflow または slice を import する
- **THEN** 品質ゲートは runtime 境界違反として報告しなければならない
