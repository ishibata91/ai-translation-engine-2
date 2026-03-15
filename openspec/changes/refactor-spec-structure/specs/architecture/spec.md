## ADDED Requirements

### Requirement: architecture は spec 区分の参照導線を実装責務区分へ一致させなければならない
システムは、`architecture` 文書の関連文書参照において `governance`、`frontend`、`controller`、`workflow`、`slice`、`runtime`、`artifact`、`gateway`、`foundation` の区分を曖昧化してはならない。特に `artifact` と `foundation` は独立した区分として参照されなければならない。

#### Scenario: 開発者が architecture から関連 spec を辿る
- **WHEN** 開発者が `architecture` 文書の関連文書一覧を見る
- **THEN** `artifact` と `foundation` を含む正規区分に整合する参照先を辿れなければならない
- **AND** `cross-cutting` のような曖昧な分類へ戻してはならない

### Requirement: architecture は governance 文書への委譲先を canonical path で示さなければならない
システムは、`architecture` 文書から品質ゲート、テスト設計、ログ運用、frontend 構造、spec 配置ルールを参照するとき、`openspec/specs/<zone>/<capability>/...` の canonical path で案内しなければならない。

#### Scenario: architecture が品質ゲートやログ設計を参照する
- **WHEN** `architecture` 文書が自身の責務外である運用基準を案内する
- **THEN** 参照先は `governance` または `frontend` の canonical path で示されなければならない
- **AND** root 直下の旧パスを正本として案内してはならない
