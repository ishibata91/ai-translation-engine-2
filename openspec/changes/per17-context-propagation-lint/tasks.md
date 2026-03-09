## 1. 検査仕様の固定

- [x] 1.1 `backend-quality-gates` delta spec に `context` 伝播違反の検出対象と除外境界を反映する
- [x] 1.2 公開入口、`context.Background()` / `TODO()`、goroutine 起点の代表違反ケースを fixture 一覧として整理する

## 2. Analyzer 実装

- [x] 2.1 `tools/backendquality` に `go/analysis` ベースの `context` 伝播 analyzer 実行層を追加する
- [x] 2.2 `backend:lint` から既存 lint 後に analyzer を実行し、違反を統一フォーマットで報告する
- [x] 2.3 composition root やテスト補助などの正当除外ケースを誤検知しないフィルタを実装する

## 3. 検証と導入

- [x] 3.1 analyzer fixture / テストで未伝播、`Background`、goroutine 脱落を再現し、検出結果を固定する
- [x] 3.2 `npm run backend:lint` と必要に応じて `npm run backend:check` を実行し、導入影響を確認する
- [x] 3.3 既存コードに検出された違反の扱いを整理し、別修正または段階導入方針を記録する
