# Change Review: standardize-e2e-required-scenarios

## 1. ユーザが出した完了条件

- 完了条件 1: ページ単位 E2E の `必須シナリオ` 標準を main specs に反映する
- 完了条件 2: `DictionaryBuilder` の必須シナリオ 3 本を PageObject 中心で実装する
- 完了条件 3: frontend 品質ゲート（`lint:file` / `lint:frontend` / `e2e`）を通す

## 2. 品質ゲート確認

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
  - `npm run lint:file -- src/e2e/helpers/wails-mock.ts src/e2e/fixtures/dictionary-builder/mock-data.ts src/e2e/page-objects/pages/dictionary-builder.po.ts src/e2e/dictionary-builder-required-scenarios.spec.ts`
  - `npm run lint:file -- src/e2e/helpers/wails-mock.ts src/e2e/fixtures/dictionary-builder/mock-data.ts src/e2e/page-objects/pages/dictionary-builder.po.ts src/e2e/dictionary-builder-required-scenarios.spec.ts` (再実行)
  - `npm run lint:frontend`
  - `npm run e2e`
- 未実行の品質ゲートと理由:
  - backend 系品質ゲートは今回 backend 変更がないため未実行
- レビュー時の補足:
  - Playwright 実行結果は `7 passed (5.3s)` で、`DictionaryBuilder` 必須シナリオ 3 本を含めて通過した
