## 1. ルール定義の固定

- [x] 1.1 `backend-quality-gates` delta spec に `slog` 利用規約の検出対象と許容境界を反映する
- [x] 1.2 直接 `slog` 呼び出し、`*Context` 未使用、key 命名違反、wrapper 許容の代表ケースを fixture 観点で整理する

## 2. Analyzer 実装

- [x] 2.1 `tools/backendquality` に `slog` 利用規約 analyzer 実行層を追加する
- [x] 2.2 `slog.Info/Error/Warn/Debug` の直接利用と `*Context` 未使用を検出する MVP ルールを実装する
- [x] 2.3 文字列リテラル key に対する lower_snake_case 判定と wrapper 許容ロジックを実装する

## 3. 検証と導入

- [x] 3.1 analyzer fixture / テストで禁止パターンと許容パターンの両方を回帰確認する
- [x] 3.2 `npm run backend:lint` と必要に応じて `npm run backend:check` を実行し、導入影響を確認する
- [x] 3.3 既存 logging 呼び出しへの適用方針と残課題を記録する
