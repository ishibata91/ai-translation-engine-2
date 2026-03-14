## MODIFIED Requirements

### Requirement: 実行導線の定義
システムは、ローカルとCIの双方で同一品質ゲートを実行できる導線を提供しなければならない。

#### Scenario: ローカルとCIで同じ判定基準になる
- **WHEN** 開発者がローカル実行したチェックとCI実行結果を比較する
- **THEN** 同一コマンド群または同等設定で判定結果が一致する

#### Scenario: controller API テストが backend test 導線に含まれる
- **WHEN** 開発者が `go test ./pkg/...`、`npm run backend:test`、または `npm run backend:check` を実行する
- **THEN** `pkg/controller/**` の API テストが既存の `pkg` テストと同じ導線で実行される
- **AND** controller API テストだけを特別な別手順なしで日常確認に含められる
