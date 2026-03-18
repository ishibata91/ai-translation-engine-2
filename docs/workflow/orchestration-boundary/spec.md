# workflow-orchestration-boundary Specification

## Purpose
TBD - created by archiving change fix-workflow-boundary-dependencies. Update Purpose after archive.
## Requirements
### Requirement: workflowはorchestrationに専念しなければならない
システムは、`pkg/workflow/**` を slice と runtime を束ねる orchestration に専念させなければならない。workflow は `controller`、`gateway`、`artifact` の実装詳細を直接保持してはならない。

#### Scenario: workflowが実行制御とslice呼び分けを束ねる
- **WHEN** 開発者が複数 contract を束ねるユースケース進行を実装する
- **THEN** workflow は slice 呼び分け、runtime 利用、artifact 識別子管理に集中しなければならない
- **AND** 外部 I/O の具象実装や UI 境界処理を直接内包してはならない

### Requirement: workflowは共有データ本体ではなくartifact識別子を管理しなければならない
システムは、slice 間で受け渡す共有データ本体を workflow 内へ保持し続けるのではなく、`artifact` 識別子、検索条件、batch / page / cursor を管理しなければならない。

#### Scenario: workflowが後続sliceへ共有データを渡す
- **WHEN** workflow が前段 slice の成果物を後続 slice へ引き渡す
- **THEN** workflow は artifact 識別子または検索条件を束ねて後続 slice を呼び出さなければならない
- **AND** workflow は gateway 実装や slice 内部保存物を直接中継してはならない

### Requirement: workflowテストも同じ責務境界に従わなければならない
システムは、`pkg/workflow/**` 配下の test code においても workflow の責務境界を維持しなければならない。controller や gateway を直接 import して責務違反を補強してはならない。

#### Scenario: workflow配下のテストで境界違反を追加する
- **WHEN** 開発者が `pkg/workflow/**` 配下の test code で controller または gateway への直接依存を追加する
- **THEN** 品質ゲートは workflow 境界違反として報告しなければならない

