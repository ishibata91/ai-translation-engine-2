# Purpose

slice 間で受け渡す共有データ、中間成果物、resume 用状態の保存・検索境界を定義し、workflow 主導の handoff を成立させる。

## Requirements

### Requirement: slice間共有データはartifactへ配置しなければならない
システムは、複数 slice から参照される共有データ、中間成果物、resume 用状態を `artifact` に配置しなければならない。ある slice の内部保存物を、後続 slice が直接参照してはならない。

#### Scenario: 後続sliceが前段sliceの成果物を利用する
- **WHEN** ある slice が後続 slice へ渡す共有データを保存する必要がある
- **THEN** 共有データは `artifact` の保存・検索契約へ格納されなければならない
- **AND** 後続 slice は前段 slice の内部 DB や内部 DTO を直接参照してはならない

### Requirement: slice間受け渡しはworkflowがartifact境界で束ねなければならない
システムは、slice 間の受け渡しを `workflow` が `artifact` 識別子、検索条件、batch / page / cursor を用いて束ねなければならない。

#### Scenario: slice間連携を実装する
- **WHEN** 開発者が parser の出力を persona や translator へ受け渡す処理を実装する
- **THEN** `workflow` が artifact 上の識別子または検索条件を束ねて後続 slice を呼び出さなければならない
- **AND** artifact は保存・検索以外の業務判断を持ってはならない
