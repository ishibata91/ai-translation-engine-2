# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここには change 固有の観点だけを書く。共通観点は `openspec/review_standard.md` を前提とする。

## 1. Scope

- この change で主に確認したい領域: `pkg/task`, `pkg/workflow`, `pkg/config`, `pkg/pipeline`
- 影響を強く受ける package / feature: Wails バインディング経由のコンテキスト伝播、task 実行管理、pipeline 永続化、設定保存
- 特に見落としやすい境界: controller -> workflow -> task/store の `context.Context` 受け渡し、DB 層の error wrap、ログの `slog.*Context` 統一

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

- [ ] `context.Context` を追加・伝播した影響で Wails バインディング互換が崩れていないか
- [ ] error wrap 追加後も既存テストや UI 表示が期待どおりに動くか
- [ ] task / pipeline のログ・状態更新経路を整理した際に resume/cancel の挙動が変わっていないか

## 5. Verification Notes

- 重点的に見るべきファイル: `pkg/task/bridge.go`, `pkg/task/manager.go`, `pkg/workflow/master_persona_service.go`, `pkg/config/config_service.go`, `pkg/pipeline/store.go`, `pkg/pipeline/manager.go`
- 重点的に見るべきテスト: task resume/cancel 系、config service/store 系、pipeline store/manager 系
- verify 時の補足メモ: `npm run backend:lint:file -- <file...>` を対象ごとに回し、最後に `npm run lint:backend` と必要テストで収束確認する
