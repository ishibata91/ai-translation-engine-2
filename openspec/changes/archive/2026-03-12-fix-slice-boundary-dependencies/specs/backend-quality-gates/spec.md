## MODIFIED Requirements

### Requirement: 依存方向 lint を品質ゲートへ含めなければならない
システムは `depguard` を用いて `architecture.md` が定義する責務区分の import 依存方向違反を検出し、バックエンド品質ゲートへ含めなければならない。

- `controller` は `workflow` 以外の `slice`、`runtime`、`artifact`、`gateway` を直接 import してはならない。
- `workflow` は `controller`、`artifact`、`gateway` を直接 import してはならない。
- `usecase slice` は `artifact` 以外の `controller`、`workflow`、`runtime`、`gateway`、他 `slice` を直接 import してはならない。
- `runtime` は `gateway` 以外の `controller`、`workflow`、`slice`、`artifact` を直接 import してはならない。
- `artifact` は `controller`、`workflow`、`runtime`、`slice`、`gateway` を直接 import してはならない。
- `gateway` は `controller`、`workflow`、`runtime`、`slice`、`artifact` を直接 import してはならない。
- `depguard` は `pkg/**` の本番コードだけでなく、同じ責務区分配下のテストコードにも適用されなければならない。

#### Scenario: 依存方向違反が検出される
- **WHEN** controller が runtime 具象へ直接依存するなどの違反を追加する
- **THEN** `depguard` は違反を検出しなければならない

#### Scenario: runtime から gateway の限定依存だけが許可される
- **WHEN** queue worker が LLM gateway を利用する
- **THEN** 品質ゲートは当該依存を許可しなければならない
- **AND** runtime から slice 具象への依存は許可してはならない

#### Scenario: slice はartifact以外の他区分へ依存しない
- **WHEN** 開発者が slice から workflow、runtime、gateway、他 slice への import を追加する
- **THEN** `depguard` は違反を検出しなければならない
- **AND** slice から `artifact` への依存だけを許可しなければならない

#### Scenario: slice配下のテストも同じ境界で検査される
- **WHEN** 開発者が `pkg/slice/**` 配下の test code で workflow、runtime、gateway、他 slice を import する
- **THEN** `depguard` は本番コードと同じ slice 境界違反として報告しなければならない

#### Scenario: package区分ごとに対応するルールだけが適用される
- **WHEN** 開発者が `pkg/controller/**`、`pkg/workflow/**`、`pkg/slice/**`、`pkg/runtime/**`、`pkg/artifact/**`、`pkg/gateway/**` のいずれかで lint を実行する
- **THEN** `depguard` は当該 package 区分に対応する依存方向ルールだけを適用しなければならない
- **AND** 他区分向けルールを全 package に無差別適用して誤検知を増やしてはならない
