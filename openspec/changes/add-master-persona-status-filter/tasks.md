## 1. Persona ストアの状態管理

- [ ] 1.1 `pkg/persona/store.go` の `npc_personas` スキーマに `status` カラムを追加し、初期値を `draft` として定義する
- [ ] 1.2 `SavePersonaBase` で下書き保存時に `status='draft'` を設定し、`SavePersona` で保存成功時に `status='generated'` へ更新する
- [ ] 1.3 Persona ストアのテストを更新し、リクエスト生成時は `draft`、保存完了時は `generated` になることを検証する

## 2. 一覧DTOとサービス応答の更新

- [ ] 2.1 `pkg/persona/dto.go` と `ListNPCs` クエリを更新し、一覧DTOに `status` を含めて返せるようにする
- [ ] 2.2 一覧取得まわりのテストを更新し、関連ダイアログ件数と `status` が同時に返ることを確認する

## 3. MasterPersona UI の表示とフィルタ

- [ ] 3.1 `frontend/src/types/npc.ts` に `draft` / `generated` を扱う状態型、表示ラベル、バッジ定義を追加する
- [ ] 3.2 `frontend/src/pages/MasterPersona.tsx` で一覧行の固定 `完了` マッピングを廃止し、API 応答の `status` から `下書き` / `生成済み` を表示する
- [ ] 3.3 `frontend/src/pages/MasterPersona.tsx` にステータスフィルタ UI を追加し、既存の検索語・プラグイン絞り込みと併用できるようにする

## 4. 仕様同期と検証

- [ ] 4.1 [openspec/specs/database_erd.md](C:/Users/shiba/.codex/worktrees/75c7/ai%20translation%20engine%202/openspec/specs/database_erd.md) の persona ER 図に `npc_personas.status` を反映する
- [ ] 4.2 バックエンドテストと必要なフロントエンド検証を実行し、一覧の状態表示とフィルタが仕様どおり動作することを確認する
