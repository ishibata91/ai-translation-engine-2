## ADDED Requirements

### Requirement: LLM 実行は task 種別非依存で LM Studio 設定を解決しなければならない
`llm` は特定スライスの知識を持たず、Queue worker から渡される task 実行コンテキストに対して `config` から `provider/model` を再読込して実行しなければならない。`provider` が `lmstudio` 以外の場合は実行を開始してはならない。

#### Scenario: 再開時に最新 config の model で実行される
- **WHEN** Queue worker が request を再開する
- **THEN** `llm` は再開時点の `config` に保存された `model` を使って LM Studio 呼び出しを実行しなければならない

#### Scenario: 再開メタデータ欠損時は失敗する
- **WHEN** request 再開時に `config` から `provider` または `model` を取得できない
- **THEN** `llm` は再開不可エラーを返し、実行を開始してはならない
