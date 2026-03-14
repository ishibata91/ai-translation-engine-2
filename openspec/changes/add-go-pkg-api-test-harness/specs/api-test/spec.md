## ADDED Requirements

### Requirement: controller API テスト対象の定義
システムは、API テスト基盤の対象を `pkg/controller/**` の公開メソッドに限定して定義しなければならない。

#### Scenario: API テスト対象が controller 入口に固定される
- **WHEN** 開発者が API テスト基盤の対象範囲を確認する
- **THEN** 対象は `pkg/controller/**` の公開メソッドとして明記されている
- **AND** `pkg/workflow/**`、`pkg/slice/**`、`pkg/runtime/**` は API テスト基盤の対象外として扱われている

### Requirement: controller API テストの共通セットアップ提供
システムは、controller API テストで再利用できる共通セットアップを提供しなければならない。共通セットアップは少なくとも controller 初期化、必要な依存のテストダブルまたはインメモリ実装、`context.Context` と logger の準備を含まなければならない。

#### Scenario: controller ごとに同じ初期化を重複しない
- **WHEN** 開発者が複数の controller API テストを追加する
- **THEN** DB、config store、logger、context 準備を各テストファイルへ重複実装せずに利用できる

### Requirement: controller API テストは table-driven 方針に従う
システムは、controller API テストを `standard_test_spec.md` に整合する table-driven test 中心で実装できなければならない。

#### Scenario: 公開メソッドの正常系と異常系を表形式で追加できる
- **WHEN** 開発者が controller の公開メソッドに対する正常系と異常系を追加する
- **THEN** テストケースは入力、前提状態、期待結果を持つ table-driven test として整理できる
- **AND** 失敗時の調査に必要な `context.Context` と構造化ログ前提を維持できる

### Requirement: controller API テストの実行導線を提供
システムは、controller API テストを既存のバックエンド品質確認導線で実行できなければならない。少なくとも `go test ./pkg/...`、`npm run backend:test`、`npm run backend:check` から実行対象に含まれなければならない。

#### Scenario: 日常のバックエンド確認で controller API テストが実行される
- **WHEN** 開発者が `go test ./pkg/...` または `npm run backend:test` を実行する
- **THEN** controller API テストが既存の `pkg` テストと同じ導線で実行される

#### Scenario: lint と test の確認順に自然に組み込まれる
- **WHEN** 開発者が backend 変更時の標準フローに従う
- **THEN** `backend:lint:file -> 修正 -> 再実行 -> lint:backend -> backend:test` の流れで controller API テストまで確認できる
