## ADDED Requirements

### Requirement: config 関連責務は store / access / workflow 解釈へ分離されなければならない
システムは、設定関連責務を `gateway` の保存 adapter、`runtime` の実行時読取補助、`workflow` の固有解釈へ分離しなければならない。

#### Scenario: LLM 周辺の config 責務を分離する
- **WHEN** 開発者が LLM 周辺の設定読み書き実装を整理する
- **THEN** store contract と SQLite 実装は gateway 側に置かれなければならない
- **AND** TypedAccessor のような実行時読取補助は runtime 側に置かれなければならない
- **AND** prompt default のような workflow 固有解釈は workflow 側に置かれなければならない
