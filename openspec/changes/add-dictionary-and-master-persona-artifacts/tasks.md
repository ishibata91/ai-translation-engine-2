## 1. Artifact Schema and Contracts

- [ ] 1.1 `openspec/specs/governance/database-erd/spec.md` に shared dictionary artifact と master persona final / temp artifact の正本テーブル定義を追加する
- [ ] 1.2 `pkg/artifact/dictionary_artifact` の DTO、repository contract、migration を追加する
- [ ] 1.3 `pkg/artifact/master_persona_artifact` の final / temp DTO、repository contract、migration を追加する
- [ ] 1.4 `pkg/artifact/master_persona_artifact` に `source_plugin + speaker_id` lookup key DTO を定義する

## 2. Dictionary Artifact Integration

- [ ] 2.1 既存 `pkg/slice/dictionary/store.go` の shared dictionary SQL と CRUD を `pkg/artifact/dictionary_artifact` へ移す
- [ ] 2.2 dictionary slice の service / controller から artifact repository を使う DI 配線へ差し替える
- [ ] 2.3 source 単位一覧、source 作成 / 削除、entry 追加 / 更新 / 削除、横断検索を artifact 正本で成立させる

## 3. Master Persona Artifact Persistence

- [ ] 3.1 generated persona の final 保存、一覧取得、詳細取得、lookup を `pkg/artifact/master_persona_artifact` に実装する
- [ ] 3.2 下書き、生成リクエスト準備、resume 用の task 中間生成物を temp テーブルへ保存する
- [ ] 3.3 final 成果物の保存項目を `persona_id`、`form_id`、`source_plugin`、`speaker_id`、`npc_name`、`editor_id`、`race`、`sex`、`voice_type`、`updated_at`、`persona_text`、`generation_request`、`dialogues` に揃える
- [ ] 3.4 final 成果物と persona UI DTO から `status`、`dialogueCount`、`dialogue_count_snapshot` を除去する

## 4. Persona Slice Integration

- [ ] 4.1 `pkg/slice/persona` の保存、一覧、詳細、lookup を artifact repository ベースへ差し替える
- [ ] 4.2 persona slice の list / detail DTO を generated-only の final 成果物前提へ更新する
- [ ] 4.3 `generation_request` と `dialogues` を final 成果物由来で返すようにする

## 5. Workflow Cleanup Boundary

- [ ] 5.1 workflow から artifact 直接参照を追加せず、dictionary / persona slice 契約だけを呼ぶ形に揃える
- [ ] 5.2 MasterPersona task の terminal state で persona slice cleanup を呼び出す
- [ ] 5.3 cleanup が task スコープ temp だけを削除し、final 成果物を残すことを確認する

## 6. Frontend Alignment

- [ ] 6.1 `frontend/src/pages/MasterPersona.tsx` の一覧列から `status` と `dialogueCount` を削除する
- [ ] 6.2 `frontend/src/pages/MasterPersona.tsx` の status filter UI と関連 state を削除する
- [ ] 6.3 `frontend/src/components/PersonaDetail.tsx` のメタ情報から dialogue count 表示を削除する
- [ ] 6.4 `frontend/src/types/npc.ts` と関連 hook / mapper を final 成果物 DTO に合わせる

## 7. Backend Verification

- [ ] 7.1 変更した Go ファイルごとに `npm run backend:lint:file -- <file...>` を実行する
- [ ] 7.2 lint 修正後に対象ファイルで `npm run backend:lint:file -- <file...>` を再実行する
- [ ] 7.3 artifact / slice / workflow 変更に対する `go test ./pkg/...` を実行する
- [ ] 7.4 最終確認として `npm run lint:backend` を実行する

## 8. Frontend Verification

- [ ] 8.1 変更したフロントファイルごとに `npm run lint:file -- <file...>` を実行する
- [ ] 8.2 lint 修正後に対象ファイルで `npm run lint:file -- <file...>` を再実行する
- [ ] 8.3 最終確認として `npm run typecheck` を実行する
- [ ] 8.4 最終確認として `npm run lint:frontend` を実行する
- [ ] 8.5 最終確認として Playwright E2E を実行する
