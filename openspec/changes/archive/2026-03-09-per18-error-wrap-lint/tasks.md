## 1. ルール境界の定義

- [x] 1.1 `backend-quality-gates` delta spec に wrap 必須境界と許容例外を反映する
- [x] 1.2 package 境界 `return err`、`fmt.Errorf` `%w` 欠落、cleanup 例外の代表ケースを fixture 観点で整理する

## 2. Analyzer 実装

- [x] 2.1 `tools/backendquality` に error wrap analyzer 実行層を追加する
- [x] 2.2 `return err`、`fmt.Errorf`、error 握りつぶしを判定する MVP ルールを実装する
- [x] 2.3 cleanup / best-effort 処理を本流違反と分離する例外ロジックを実装する

## 3. 検証と導入

- [x] 3.1 analyzer fixture / テストで wrap 不足と許容例外の両方を回帰確認する
- [x] 3.2 `npm run backend:lint` と必要に応じて `npm run backend:check` を実行し、導入影響を確認する
- [x] 3.3 既存コードへの適用方針と残課題を記録する
