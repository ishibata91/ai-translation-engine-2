## ADDED Requirements

### Requirement: 品質ゲートの必須ツール定義
システムは、バックエンド品質ゲートの必須ツールとして `golangci-lint`、`goimports`、`goleak` を定義しなければならない。

#### Scenario: 必須ツールセットが明示される
- **WHEN** 品質ゲート定義を確認する
- **THEN** `golangci-lint`、`goimports`、`goleak` が必須実行項目として列挙されている

### Requirement: lintルールセットの標準化
システムは、`golangci-lint` で実行する標準lintセットを定義しなければならない。標準lintセットは少なくとも `staticcheck`、`govet`、`errcheck`、`revive`、`gosec` を含まなければならない。

#### Scenario: lint実行結果の一貫性が確保される
- **WHEN** 開発者が同一設定で lint を実行する
- **THEN** 実行対象lintとfail条件が統一され、環境差分による判定揺れが発生しない

### Requirement: 実行導線の定義
システムは、ローカルとCIの双方で同一品質ゲートを実行できる導線を提供しなければならない。

#### Scenario: ローカルとCIで同じ判定基準になる
- **WHEN** 開発者がローカル実行したチェックとCI実行結果を比較する
- **THEN** 同一コマンド群または同等設定で判定結果が一致する

### Requirement: lint設定の専用テストを要求しない
システムは、lint設定そのものに対する専用テストコードの作成を必須要件にしてはならない。lintは設定ファイルと実行コマンドにより運用しなければならない。

#### Scenario: lint運用がシンプルに維持される
- **WHEN** 開発者が品質ゲート要件を確認する
- **THEN** lint設定の単体テストや専用テストケース作成は必須化されていない

### Requirement: 脆弱性検査の任意運用
システムは、`govulncheck` を常時必須ゲートに含めず、依存更新時またはリリース前の任意検査として定義しなければならない。

#### Scenario: ローカル専用運用に合わせた検査強度を維持する
- **WHEN** 品質ゲートを日常開発で実行する
- **THEN** `govulncheck` は必須失敗条件に含まれず、任意実行手順として別途定義されている
