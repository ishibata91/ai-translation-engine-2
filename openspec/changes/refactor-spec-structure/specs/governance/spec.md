## ADDED Requirements

### Requirement: Governance 区分は全体基準文書の正規配置先でなければならない
システムは、`architecture`、`spec-structure`、品質ゲート、テスト標準、ログ標準、全体要求のような全区分向け文書を `governance` 区分へ配置しなければならない。これらの文書を `frontend`、`workflow`、`slice`、`runtime`、`artifact`、`gateway`、`foundation` のいずれかの capability 文書へ混在させてはならない。

#### Scenario: repo 全体の規約文書を追加する
- **WHEN** 開発者が複数責務区分にまたがる共通規約を追加する
- **THEN** 当該文書は `openspec/specs/governance/` 配下へ配置されなければならない
- **AND** 特定 capability の `spec.md` に抱え込んではならない

### Requirement: Governance 文書は区分別参照導線を提供しなければならない
システムは、`governance` 配下の文書から `frontend`、`controller`、`workflow`、`slice`、`runtime`、`artifact`、`gateway`、`foundation` の各区分を参照できる導線を提供しなければならない。`governance` 文書は実装責務そのものを代替してはならない。

#### Scenario: AI が設計時の参照先を判断する
- **WHEN** AI または開発者が共通基準文書から実装責務の詳細 spec を辿る
- **THEN** `governance` 文書から該当区分の canonical spec を一意に辿れなければならない
- **AND** `governance` 文書自体が個別 capability の詳細要件を代替してはならない
