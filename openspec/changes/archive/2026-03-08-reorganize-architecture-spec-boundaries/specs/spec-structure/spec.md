## ADDED Requirements

### Requirement: architecture.md は純粋なアーキテクチャ文書でなければならない
`architecture.md` は、パッケージ責務、依存方向、DTO / Contract / DI 原則、composition root の責務だけを定義しなければならない。品質ゲート、テスト設計、ログ運用、フロント固有構造を重複して記述してはならない。

#### Scenario: architecture.md が構造説明に集中する
- **WHEN** 開発者が `architecture.md` を参照する
- **THEN** その文書から責務区分と依存方向を判断できなければならない
- **AND** 品質ゲートやテスト設計の詳細は専用 spec への参照に委譲されていなければならない

### Requirement: 共通要件はユースケース spec から分離できなければならない
UI、workflow、runtime、gateway など複数ユースケースで共有される要件は、ユースケース単位 spec に埋め込まず、共通 spec として分離できる構成でなければならない。

#### Scenario: 共通要件が専用 spec へ切り出される
- **WHEN** UI 要件や runtime 要件が複数ユースケースで重複している
- **THEN** システムは当該要件を共通 spec として切り出せなければならない
- **AND** ユースケース spec は固有要件だけに集中できなければならない

### Requirement: AGENTS.md は文書責務の入口を示さなければならない
`AGENTS.md` は、設計・提案・実装時にどの spec を参照するべきかを文書責務ごとに示さなければならない。

#### Scenario: AI が適切な spec を参照できる
- **WHEN** AI がアーキテクチャ、品質ゲート、テスト設計、ログ運用のいずれかを検討する
- **THEN** `AGENTS.md` から該当 spec を一意に辿れなければならない
