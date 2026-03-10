## ADDED Requirements

### Requirement: gatewayは外部資源への依頼口に専念しなければならない
システムは、`pkg/gateway/**` を LLM、外部 API、file、secret、config など外部資源への依頼口に専念させなければならない。gateway は `controller`、`workflow`、`runtime`、`slice`、`artifact` を直接 import してはならず、上位層の進行制御や状態解釈を持ってはならない。

#### Scenario: gatewayが技術接続を提供する
- **WHEN** 開発者が外部 API やファイル I/O の技術接続を実装する
- **THEN** 当該実装は gateway に置かれなければならない
- **AND** gateway は workflow や runtime の進行制御知識を持ってはならない

### Requirement: gatewayはruntimeから消費できる中立DTOを返さなければならない
システムは、gateway の返却値を runtime から消費できる中立 DTO に保たなければならない。gateway は workflow 状態や slice 保存判断を含む型を返してはならない。

#### Scenario: runtimeがgateway結果を受け取る
- **WHEN** runtime が gateway を呼び出して外部 I/O 結果を受け取る
- **THEN** gateway は技術接続結果だけを返さなければならない
- **AND** workflow や slice の都合に応じた意味解釈は返してはならない

### Requirement: gateway配下のテストも同じ責務境界に従わなければならない
システムは、`pkg/gateway/**` 配下の test code においても gateway の責務境界を維持しなければならない。workflow や runtime への直接依存をテスト都合で許してはならない。

#### Scenario: gateway配下のテストで境界違反を追加する
- **WHEN** 開発者が `pkg/gateway/**` 配下の test code で workflow または runtime を import する
- **THEN** 品質ゲートは gateway 境界違反として報告しなければならない
