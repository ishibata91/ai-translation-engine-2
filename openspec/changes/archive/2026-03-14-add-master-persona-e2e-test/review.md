# Review Checklist: add-master-persona-e2e-test

このレビューは `openspec/review_standard.md` を前提に、change 固有の完了条件と品質ゲート確認を記録する。

## 1. ユーザが出した完了条件

- 完了条件 1: `MasterPersona` の必須 5 シナリオ（初期表示 / NPC 詳細 / PromptSettingCard / ModelSettings / JSON選択→タスク開始）が spec と E2E 実装で定義されている
- 完了条件 2: Page Object 中心の責務境界を維持し、spec へ locator 実装詳細を漏らさない
- 完了条件 3: フロントエンド品質ゲート（`lint:file` / `typecheck` / `lint:frontend` / 対象 E2E）が通過している

## 2. 品質ゲート確認

該当するものだけ記入する。

### Backend

- [ ] 変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行した
- [ ] `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進めた
- [ ] 作業中または完了前に `npm run lint:backend` を実行した
- [ ] 必要に応じて `npm run backend:check` または `npm run backend:watch` で品質確認した

### Frontend

- [x] 変更中ファイルに対して `npm run lint:file -- <file...>` を逐次実行した
- [x] `lint:file -> 修正 -> 再実行 -> 最後に lint:frontend` の順で進めた
- [x] 作業完了前に `npm run lint:frontend` を実行した

## 3. 実行メモ

- 実行したコマンド:
  - `npm run lint:file -- src/e2e/master-persona-required-scenarios.spec.ts src/e2e/page-objects/pages/master-persona.po.ts src/e2e/helpers/wails-mock.ts src/e2e/fixtures/master-persona/mock-data.ts`
  - `npm run typecheck`
  - `npm run lint:frontend`
  - `npm run e2e -- src/e2e/master-persona-required-scenarios.spec.ts`
- 未実行の品質ゲートと理由:
  - Backend 系は対象外（本 change は frontend / e2e のみ）
- レビュー時の補足:
  - `MasterPersona` 必須シナリオは `frontend/src/e2e/master-persona-required-scenarios.spec.ts` で 5 本を確認。
  - Page Object は `frontend/src/e2e/page-objects/pages/master-persona.po.ts` に集約され、spec 側に locator 詳細は漏れていない。
  - delta spec (`openspec/changes/add-master-persona-e2e-test/specs/...`) と main spec (`openspec/specs/e2e-required-scenarios/master-persona/spec.md`) の要件内容は同期済み。