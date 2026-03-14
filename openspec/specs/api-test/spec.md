# Purpose

`pkg/controller/**` の公開メソッドを対象にした API テスト基盤の責務、配置、実行導線を定義する。

## Requirements

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

### Requirement: 現行 controller 群の公開 API テスト整備
システムは、`pkg/controller` に存在する現行 controller 群の公開メソッドに対して API テストを整備し、正常系と主要異常系の回帰を検知できなければならない。

#### Scenario: 既存 controller の公開メソッドが API テストで保護される
- **WHEN** 開発者が `pkg/controller` 配下の既存 controller を確認する
- **THEN** `ConfigController`、`DictionaryController`、`ModelCatalogController`、`PersonaController`、`PersonaTaskController`、`TaskController`、`FileDialogController` の公開メソッドに対応する API テストが存在する
- **AND** 各テストは controller 単位で table-driven test を基本形として整理されている

#### Scenario: controller 境界の責務を回帰検知できる
- **WHEN** 開発者が controller API テストを実行する
- **THEN** 正常系だけでなく、入力検証エラー、依存先エラー、nil 構成時のガードなど主要異常系も検証できる
- **AND** `context.Context` 伝播、error wrap、戻り値整形など controller 境界の責務が回帰検知対象になっている

#### Scenario: Dictionary Builder と Master Persona の主要導線を controller API から検証できる
- **WHEN** 開発者が Dictionary Builder または Master Persona に関わる controller API テストを実行する
- **THEN** Dictionary Builder ではインポート開始と主要な取得系 API、Master Persona では開始、再開、状態取得の主要導線を検証できる
- **AND** controller 境界の入力写像、error 伝播、主要な戻り値契約が回帰検知対象になっている

### Requirement: Dictionary Builder と Master Persona の詳細テストスコープ配置
システムは、Dictionary Builder と Master Persona の主要テスト対象を `specs/api-test/` 配下の分割文書として保持し、親 spec から辿れるようにしなければならない。

#### Scenario: Dictionary Builder の詳細スコープを分割配置する
- **WHEN** 開発者が Dictionary Builder の controller API テスト対象を確認する
- **THEN** 詳細な対象メソッドと観点は `specs/api-test/dictionary-builder.md` に整理されている
- **AND** 親 spec だけでは不足する具体ケースをその文書から参照できる

#### Scenario: Master Persona の詳細スコープを分割配置する
- **WHEN** 開発者が Master Persona の controller API テスト対象を確認する
- **THEN** 詳細な対象メソッドと観点は `specs/api-test/master-persona.md` に整理されている
- **AND** 親 spec だけでは不足する具体ケースをその文書から参照できる

### Requirement: controller 依存パターン別の API テスト builder 提供
システムは、現行 controller 群の依存差分に応じて API テスト builder を分離し、共通 `testenv` に責務を詰め込みすぎずに controller ごとのテストを追加できなければならない。

#### Scenario: service 系 controller を局所スタブで組み立てられる
- **WHEN** 開発者が `DictionaryController`、`ModelCatalogController`、`PersonaController` の API テストを追加する
- **THEN** 各 controller は `pkg/tests/api_tests/<controller>` 配下の builder または fake で必要依存だけを組み立てられる
- **AND** 共通 `testenv` には DB、logger、trace context など横断基盤だけが残る

#### Scenario: workflow と manager を持つ controller を個別 builder で扱える
- **WHEN** 開発者が `PersonaTaskController` または `TaskController` の API テストを追加する
- **THEN** manager、workflow、store の差分は controller 別 builder で吸収される
- **AND** 他 controller 向けの依存を同じ builder に混在させない

### Requirement: Wails runtime 依存 controller の API テスト seam
システムは、Wails runtime に直接依存する controller について、公開 API の契約を backend API テストで検証できる seam を提供しなければならない。

#### Scenario: FileDialogController を runtime 実行なしで検証できる
- **WHEN** 開発者が `FileDialogController` の API テストを実行する
- **THEN** ダイアログ呼び出しは差し替え可能な seam 経由で制御できる
- **AND** ファイル種別フィルタ、戻り値、error wrap を Wails 実行環境なしで検証できる
