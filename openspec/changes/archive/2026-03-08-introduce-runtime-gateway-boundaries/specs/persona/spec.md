## MODIFIED Requirements

### Requirement: 独立性: ペルソナ生成データの受け取りと独自DTO定義
本 slice は、自前の入力 DTO と保存 DTO を契約として公開しなければならない。persona slice は runtime 制御に依存してはならず、必要な外部資源は gateway 契約経由で利用しなければならない。

#### Scenario: persona slice は runtime を参照しない
- **WHEN** persona slice の contract と実装を参照する
- **THEN** queue、progress、resume など runtime 制御へ依存してはならない

#### Scenario: persona slice は gateway 契約を利用できる
- **WHEN** persona slice が永続化や LLM 依頼準備に必要な外部資源を使う
- **THEN** gateway 契約を通じて利用しなければならない
