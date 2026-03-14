## ADDED Requirements

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
