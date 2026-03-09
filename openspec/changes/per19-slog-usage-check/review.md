# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここには change 固有の観点だけを書く。共通観点は `openspec/review_standard.md` を前提とする。

## 1. Scope

- この change で主に確認したい領域: `slog.*Context` 利用規約とログ key 命名の静的検査
- 影響を強く受ける package / feature: `tools/backendquality`, `pkg/**`, 構造化ログ呼び出し
- 特に見落としやすい境界: `slog.Info/Error/Warn` 直接利用と lower_snake_case でない key

## 2. Backend-Specific Checks

必要な場合のみ記入する。

- [ ] Contract / DTO / Mapper / Orchestrator の責務境界が change の意図どおりか
- [ ] `context.Context` の伝播が新規コードでも維持されているか
- [ ] `pkg/infrastructure/telemetry` を使うべき処理で独自 logging を増やしていないか
- [ ] 外部 I/O、重要分岐、状態変化、異常系に AI 解析可能な構造化ログがあるか
- [ ] 追加した error path に文脈付き error wrap があるか
- [ ] 主要シナリオと失敗系をカバーする Go テストがあるか

## 3. Frontend-Specific Checks

必要な場合のみ記入する。

- [ ] `pages` が描画責務に専念し、Wails API や複雑な状態管理を直接持っていないか
- [ ] Wails API が `hooks/features/*` 経由に保たれているか
- [ ] `pages` から `wailsjs` / `store` を直接 import していないか
- [ ] 境界値を `unknown` + runtime parse で扱えているか
- [ ] adapter の責務が UI 層へ漏れていないか
- [ ] `lint:file -> 修正 -> 再実行 -> lint:frontend` の流れを満たしているか

## 4. Change-Specific Risks

- [ ] この change で特有の退行リスク 1: logger wrapper や alias 利用を禁止パターンとして誤認しないか
- [ ] この change で特有の退行リスク 2: key 命名チェックが定数や組み立て文字列で過剰反応しないか
- [ ] データ互換性 / 移行 / 再実行時のリスク: 既存ログ呼び出しの大量修正が品質ゲート導入コストを押し上げないか

## 5. Verification Notes

- 重点的に見るべきファイル: `tools/backendquality/**`, logging 規約 rule 実装, 対象 fixture
- 重点的に見るべきテスト: `slog.Info/Error/Warn` 直接利用, `*Context` 未使用, lower_snake_case 違反 key
- verify 時の補足メモ: 機械検査できる範囲とレビュー継続項目の線引きが明示されているか確認する
