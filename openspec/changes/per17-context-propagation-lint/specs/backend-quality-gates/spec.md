## ADDED Requirements

### Requirement: context伝播違反を品質ゲートで検出しなければならない
システムは、`pkg/**` を対象に `context.Context` の伝播違反を検出する静的検査をバックエンド品質ゲートへ含めなければならない。検査は、公開入口で受けた `ctx` の未伝播、不適切な `context.Background()` / `context.TODO()` の利用、および goroutine 起点での `ctx` 脱落を検出できなければならない。

#### Scenario: 公開入口で受けたctxを内部処理へ渡さない
- **WHEN** 開発者が `ctx context.Context` を受けた公開メソッドから、I/O または外部依頼を行う内部処理へ `ctx` を渡さずに呼び出す
- **THEN** 品質ゲートは `context` 伝播違反として報告しなければならない

#### Scenario: context.BackgroundまたはTODOを業務処理で新規生成する
- **WHEN** 開発者が公開入口や workflow / slice / gateway の業務処理内で、既存の `ctx` を無視して `context.Background()` または `context.TODO()` を利用する
- **THEN** 品質ゲートは trace / cancel 伝播を壊す違反として報告しなければならない

#### Scenario: goroutine起点でctxを落とす
- **WHEN** 開発者が goroutine を起動する処理で、親 `ctx` を引き継がずに非同期処理を開始する
- **THEN** 品質ゲートは goroutine 境界での `context` 脱落として報告しなければならない

### Requirement: context伝播検査は正当な除外境界を持たなければならない
システムは、初期化専用コード、テスト補助、純粋関数など `context.Context` 伝播義務の外にあるケースを識別し、実運用不能な誤検知を避けなければならない。

#### Scenario: 初期化専用コードは一律違反にしない
- **WHEN** 開発者が composition root やテストセットアップで `context.Background()` を利用する
- **THEN** 品質ゲートは当該コードを除外対象として扱うか、少なくとも業務処理違反とは区別して報告しなければならない
