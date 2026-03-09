# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここには change 固有の観点だけを書く。共通観点は `openspec/review_standard.md` を前提とする。

## 1. Scope

- この change で主に確認したい領域: 公開メソッドの責務過多を検出する設計 lint の成立性調査
- 影響を強く受ける package / feature: `tools/backendquality`, `pkg/**`, レビュー運用の補助基準
- 特に見落としやすい境界: 調査止まりの結論と MVP ルール化案の切り分け

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

- [ ] この change で特有の退行リスク 1: 行数やネスト深度のような単純指標に引きずられて設計品質を誤判定しないか
- [ ] この change で特有の退行リスク 2: 調査チケットであるにもかかわらず実装前提の spec を作ってしまわないか
- [ ] データ互換性 / 移行 / 再実行時のリスク: 導入見送り時にレビュー運用へ何を残すかが曖昧にならないか

## 5. Verification Notes

- 重点的に見るべきファイル: 調査メモ, `tools/backendquality/**`, 既存長大メソッドのサンプル分析
- 重点的に見るべきテスト: もし MVP を作るなら責務混在の代表ケースと誤検知ケース
- verify 時の補足メモ: 実装可否, 誤検知率, 導入コストの比較表があるか確認する
