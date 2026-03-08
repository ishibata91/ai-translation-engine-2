## ADDED Requirements

### Requirement: 共通バックエンドコーディング規約の定義
システムは、バックエンド開発で必ず従う共通規約を定義しなければならない。共通規約は少なくとも命名、エラーハンドリング、`context.Context` 伝播、構造化ログ、SRP、公開APIへのdocコメント必須を含まなければならない。

#### Scenario: 共通規約が公開される
- **WHEN** 開発者が規約ドキュメントを参照する
- **THEN** 命名、error wrap、context伝播、構造化ログ、SRP、公開API doc必須の各項目が MUST として明記されている

### Requirement: リポジトリ固有規約の定義
システムは、`architecture.md` に準拠したリポジトリ固有規約を定義しなければならない。固有規約は Interface-First AIDD、Vertical Slice、DTO分離、Pipeline Mapper責務を必須要件として含まなければならない。

#### Scenario: アーキテクチャ原則への準拠が確認できる
- **WHEN** 規約ドキュメントの固有規約セクションを確認する
- **THEN** Interface-First AIDD、Vertical Slice、DTO分離、Pipeline Mapper責務が MUST として定義されている

### Requirement: 一般的なチェックスタイルの定義
システムは、規約運用のための一般的なチェックスタイルを定義しなければならない。チェックスタイルは MUST/SHOULD の区分、レビュー時の確認観点、違反時の扱いを明示しなければならない。

#### Scenario: レビュー時の判定基準が一意になる
- **WHEN** レビュアが規約に基づいてPRを確認する
- **THEN** MUST違反とSHOULD違反の判定基準、および対応方針を同一基準で適用できる

### Requirement: テスト方針の整合
システムは、規約内のテスト方針を `standard_test_spec.md` と整合させなければならない。少なくとも Table-Driven Test 中心、`context.Context` 伝播、構造化ログ前提のデバッグフローを要求しなければならない。

#### Scenario: テスト方針の参照先が固定される
- **WHEN** 開発者がテスト方針を確認する
- **THEN** `standard_test_spec.md` を参照する旨と、Table-Driven/Context/構造化ログの要件が明示されている
