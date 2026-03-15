# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここでは `AGENTS.md` に定義された品質ゲート確認と、ユーザが提示した完了条件だけを扱う。
共通観点は `openspec/review_standard.md` を前提とする。

## 1. ユーザが出した完了条件

- 完了条件 1: `frontend/src/pages/TranslationFlow.tsx` にデータロード用の 1 フェーズが追加されている
- 完了条件 2: 複数ファイルを選択できる
- 完了条件 3: `frontend/src/components/DataTable.tsx` で翻訳対象を一覧表示し、テーブルはファイルごとに分かれて表示される
- 完了条件 4: 各ファイルのテーブルは折りたたみ可能である

## 2. 品質ゲート確認

該当するものだけ記入する。

### Backend

- [x] 変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行した
- [x] `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進めた
- [x] 作業中または完了前に `npm run lint:backend` を実行した
- [ ] 必要に応じて `npm run backend:check` または `npm run backend:watch` で品質確認した

### Frontend

- [x] 変更中ファイルに対して `npm run lint:file -- <file...>` を逐次実行した
- [x] `lint:file -> 修正 -> 再実行 -> 最後に lint:frontend` の順で進めた
- [x] 作業完了前に `npm run lint:frontend` を実行した

## 3. 実行メモ

- 実行したコマンド:
  - `npm run backend:lint:file -- <backend changed files>`
  - `npm run backend:fmt`
  - `npm run backend:lint:file -- <backend changed files>`
  - `npm run backend:lint:file -- pkg/artifact/translationinput/repository.go`
  - `npm run lint:backend`
  - `go test ./pkg/artifact/translationinput ./pkg/controller ./pkg/workflow`
  - `npm run lint:file -- <frontend changed files>`（`frontend` 配下）
  - `npm run lint:file -- src/e2e/page-objects/pages/translation-flow.po.ts src/e2e/translation-flow-required-scenarios.spec.ts`（`frontend` 配下）
  - `npm run typecheck`（`frontend` 配下）
  - `npm run lint:frontend`（`frontend` 配下）
  - `npm run e2e`（`frontend` 配下, 15 passed）
  - `go test ./pkg/...`
- 未実行の品質ゲートと理由:
  - `npm run backend:check` / `npm run backend:watch` は未実行（本変更では `lint:backend` と関連パッケージテストを優先して確認）
- レビュー時の補足:
  - `go test ./pkg/...` は既存の `pkg/gateway/llm` テスト（`TestLLMManager_GetBatchClient` の `xAI grok-3-mini` batch 非対応期待）で失敗し、本変更起因ではないことを確認
