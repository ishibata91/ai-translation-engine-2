## MODIFIED Requirements

### Requirement: 依存方向 lint を品質ゲートへ含めなければならない
システムは `depguard` を用いて `architecture.md` が定義する責務区分の import 依存方向違反を検出し、バックエンド品質ゲートへ含めなければならない。

- `foundation` は `controller`、`workflow`、`slice`、`runtime`、`gateway` のいずれにも逆依存してはならない。
- `controller`、`workflow`、`slice`、`runtime`、`gateway` は foundation を直接 import できなければならない。
- `depguard` は foundation を専用 files ルールで検査し、他区分向けルールを無差別適用してはならない。

#### Scenario: foundation への依存は許可される
- **WHEN** LLM gateway が foundation 配下の telemetry を利用する
- **THEN** 品質ゲートは当該依存を許可しなければならない
- **AND** gateway から runtime への依存は引き続き許可してはならない

#### Scenario: foundation から上位区分への逆依存は検出される
- **WHEN** foundation 配下の package が workflow や runtime の具象実装を import する
- **THEN** `depguard` は foundation 境界違反として報告しなければならない
