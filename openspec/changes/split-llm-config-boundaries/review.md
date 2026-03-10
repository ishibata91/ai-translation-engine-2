# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここには change 固有の観点だけを書く。共通観点は `openspec/review_standard.md` を前提とする。

## 1. Scope

- この change で主に確認したい領域: LLM 周辺の config 境界分離
- 影響を強く受ける package / feature: `workflow/config`、`gateway/config`、LLM manager、model catalog、queue worker、persona prompt
- 特に見落としやすい境界: `configstore` / `configaccess` / `workflow/persona` の責務混線

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

- [ ] gateway が workflow config を再エクスポートする構造が残っていないか:
- [ ] LLM 周辺限定というスコープが崩れていないか:
- [ ] データ互換性 / 移行 / 再実行時のリスク:

## 5. Verification Notes

- 重点的に見るべきファイル: `workflow/config` 分解先、LLM 周辺 import、Wire provider
- 重点的に見るべきテスト: model catalog、queue worker、persona prompt、config store 関連
- verify 時の補足メモ: 他ユースケースの config 移行を混ぜていないことを確認する
