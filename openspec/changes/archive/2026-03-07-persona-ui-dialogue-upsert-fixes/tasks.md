## 1. Personaスキーマ再定義

- [x] 1.1 `pkg/persona/store` のスキーマ初期化を更新し、`npc_personas` に `id` 主キーと `unique(source_plugin, speaker_id)` を追加する
- [x] 1.2 `pkg/persona/store` のスキーマ初期化を更新し、`npc_dialogues` に `persona_id` 外部キーを追加して原文中心カラム構成に整理する
- [x] 1.3 開発中前提として `persona.db` の既存テーブル再作成フロー（互換移行なし）を実装する
- [x] 1.4 `source_plugin` 欠損時に入力名から `*.esm|*.esl|*.esp` を補完し、不可時 `UNKNOWN` を設定する処理を実装する

## 2. Persona保存・取得ロジック更新

- [x] 2.1 `overwrite_existing` のON/OFFに応じて `npc_personas` の更新/保持を切り替える保存ロジックを実装する
- [x] 2.2 `npc_dialogues` 保存を原文中心に変更し、`translated_text` 更新依存を除去する
- [x] 2.3 `ListNPCs` の返却DTOに `persona_id` を含める
- [x] 2.4 セリフ取得APIを `ListDialoguesByPersonaID` ベースに更新する

## 3. Task入力と自動再開

- [x] 3.1 `pkg/task/master_persona_task` の `StartMasterPersonTaskInput` に `overwrite_existing` を追加する
- [x] 3.2 `overwrite_existing` を task metadata に保存し、再開時に同一方針を再利用する
- [x] 3.3 `REQUEST_GENERATED` 直後の `ResumeTask` 呼び出しを受理できる状態遷移を確認・調整する

## 4. Frontend反映

- [x] 4.1 `frontend/src/pages/MasterPersona.tsx` に「重複時上書き」チェックボックスを追加する
- [x] 4.2 タスク開始時に `overwrite_existing` を `StartMasterPersonTask` へ渡す
- [x] 4.3 `REQUEST_GENERATED` 受信時に同一 task ID で `ResumeTask` を自動実行する
- [x] 4.4 `frontend/src/components/PersonaDetail.tsx` のセリフ一覧を原文表示に統一する
- [x] 4.5 詳細読込を `persona_id` ベースに更新し、選択行と表示内容の一意性を保証する

## 5. 仕様ドキュメント同期

- [x] 5.1 `openspec/specs/database_erd.md` の persona ER を `id` 主キー + `persona_id` 参照構成へ更新する
- [x] 5.2 必要に応じて `openspec/specs/persona/spec.md` のDBスキーマ記述を実装後の仕様へ同期する

## 6. テスト・検証

- [x] 6.1 `pkg/persona` のテストに `source_plugin + speaker_id` 衝突ケース（上書きON/OFF）を追加する
- [x] 6.2 `pkg/task` のテストに `overwrite_existing` metadata 維持と再開時適用ケースを追加する
- [x] 6.3 `frontend` 側の動作確認として `REQUEST_GENERATED -> 自動ResumeTask` の遷移を検証する
- [ ] 6.4 `MasterPersona` 画面で「原文のみ表示」「重複時上書き」「再開動作」の回帰確認を実施する
