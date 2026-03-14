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

### Requirement: controller API テスト補助の配置
システムは、controller API テスト用の `testenv` と helper を `pkg/tests/api_tests` 配下へ配置しなければならない。

#### Scenario: API テスト補助の責務が集約される
- **WHEN** 開発者が controller API テスト用の helper 配置を確認する
- **THEN** 共通 `testenv` と controller 別 builder は `pkg/tests/api_tests` 配下に置かれている
- **AND** `pkg/controller/**` 直下へ汎用 helper を分散配置しない

### Requirement: controller API テストは table-driven 方針に従う
システムは、controller API テストを `standard_test_spec.md` に整合する table-driven test 中心で実装できなければならない。

#### Scenario: 公開メソッドの正常系と異常系を表形式で追加できる
- **WHEN** 開発者が controller の公開メソッドに対する正常系と異常系を追加する
- **THEN** テストケースは入力、前提状態、期待結果を持つ table-driven test として整理できる
- **AND** 失敗時の調査に必要な `context.Context` と構造化ログ前提を維持できる

### Requirement: testify を API テスト標準ライブラリとして利用する
システムは、controller API テストのアサーションとテスト補助に `stretchr/testify` を利用しなければならない。

#### Scenario: API テストのアサーション記法が統一される
- **WHEN** 開発者が controller API テストを追加する
- **THEN** `require` または `assert` を用いて失敗条件と期待値を記述できる
- **AND** 標準 `testing` だけに依存した独自アサーション補助を増やさない

### Requirement: テスト用 SQLite DB は通常 DB と別パスを利用する
システムは、controller API テストで使う SQLite DB を通常の `db/` とは別パスへ生成できなければならない。テスト用 DB の保存先は `tmp/api_test_db/` とし、バージョン管理対象外でなければならない。

#### Scenario: テスト用 DB が通常 DB と混在しない
- **WHEN** 開発者が controller API テストをファイルベース SQLite で実行する
- **THEN** テスト用 DB は `tmp/api_test_db/` 配下へ生成される
- **AND** 通常の `db/` 配下の永続 DB を汚染しない
- **AND** `tmp/api_test_db/` は `.gitignore` 対象になっている

### Requirement: controller ごとの差分は builder で吸収する
システムは、controller ごとに異なる依存構成を共通 `testenv` に詰め込みすぎず、controller 別 builder または env 作成関数で吸収しなければならない。

#### Scenario: controller 固有依存が共通 testenv を肥大化させない
- **WHEN** `ConfigController` と別 controller で必要依存が異なる
- **THEN** 共通 `testenv` は DB・logger・context・共通 utility に留まる
- **AND** controller 固有の workflow や store 構成は controller 別 builder で組み立てられる

### Requirement: controller API テストの実行導線を提供
システムは、controller API テストを既存のバックエンド品質確認導線で実行できなければならない。少なくとも `go test ./pkg/...`、`npm run backend:test`、`npm run backend:check` から実行対象に含まれなければならない。

#### Scenario: 日常のバックエンド確認で controller API テストが実行される
- **WHEN** 開発者が `go test ./pkg/...` または `npm run backend:test` を実行する
- **THEN** controller API テストが既存の `pkg` テストと同じ導線で実行される

#### Scenario: lint と test の確認順に自然に組み込まれる
- **WHEN** 開発者が backend 変更時の標準フローに従う
- **THEN** `backend:lint:file -> 修正 -> 再実行 -> lint:backend -> backend:test` の流れで controller API テストまで確認できる
