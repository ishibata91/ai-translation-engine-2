## ADDED Requirements

### Requirement: 依存方向 lint を品質ゲートへ含めなければならない
システムは `go-cleanarch` を用いて責務区分の依存方向違反を検出し、バックエンド品質ゲートへ含めなければならない。

#### Scenario: 依存方向違反が検出される
- **WHEN** controller が runtime 具象へ直接依存するなどの違反を追加する
- **THEN** `go-cleanarch` は違反を検出しなければならない

#### Scenario: runtime から gateway の限定依存だけが許可される
- **WHEN** queue worker が LLM gateway を利用する
- **THEN** 品質ゲートは当該依存を許可しなければならない
- **AND** runtime から slice 具象への依存は許可してはならない
