# Change Review Checklist

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここでは `AGENTS.md` に定義された品質ゲート確認と、ユーザが提示した完了条件だけを扱う。
共通観点は `openspec/review_standard.md` を前提とする。

## 1. ユーザが出した完了条件

- 仕様文との不一致（dictionary source の `description` / `updated_at`）は仕様書側を実装に合わせて修正する
- 仕様語彙の不一致（`translated_text` vs `dest_text`）は `dest_text` に統一する
- change 固有 review checklist を更新し、今回の品質ゲート実行結果を記録する

## 2. 品質ゲート確認

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
  - `go test ./pkg/...`
  - `npm run backend:lint:file -- pkg/artifact/translationinput/migration.go pkg/artifact/dictionary_artifact/migration.go pkg/artifact/master_persona_artifact/migration.go`
  - `npm run lint:backend`
  - `frontend: npm run lint:file -- src/components/PersonaDetail.tsx src/hooks/features/masterPersona/helpers.ts src/hooks/features/masterPersona/types.ts src/hooks/features/masterPersona/useMasterPersona.tsx src/pages/MasterPersona.tsx src/types/npc.ts`
  - `frontend: npm run typecheck`
  - `frontend: npm run lint:frontend`
  - `frontend: npm run e2e`
- 未実行の品質ゲートと理由:
  - `backend:check` / `backend:watch` は今回の修正範囲（仕様文書 + migration 軽微調整）では追加実行を省略
- レビュー時の補足:
  - `schema_version` はローカル運用前提のため、artifact migration はバージョン固定 `1` とし、テーブル作成を `CREATE TABLE IF NOT EXISTS` で常時試行する運用へ統一
