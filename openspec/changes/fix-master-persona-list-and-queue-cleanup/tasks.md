## 1. Persona データ契約とスキーマの修正

- [ ] 1.1 `pkg/persona` の保存スキーマと DTO を更新し、初回の `llm_queue` job 登録前に生成リクエストを保存できるようにする
- [ ] 1.2 `npc_personas.dialogue_count` 列を削除し、関連する保存処理とクエリから参照を除去する
- [ ] 1.3 ペルソナ一覧取得を `npc_dialogues` 件数集計ベースへ変更する
- [ ] 1.4 ペルソナ詳細取得を `RAW response` ではなく `生成リクエスト` を返す契約へ更新する

## 2. MasterPersona UI の再取得と表示修正

- [ ] 2.1 `frontend/src/pages/MasterPersona.tsx` の一覧ロード処理を共通化し、初回表示・画面復帰・関連 task イベント受信後に全件再取得する
- [ ] 2.2 ペルソナ一覧のセリフ数表示を新しい集計結果へ切り替え、旧カラム前提の表示を削除する
- [ ] 2.3 ペルソナ詳細 UI の `RAW response` 表示を `生成リクエスト` 表示へ差し替える

## 3. Task 完了時の queue クリーンアップ

- [ ] 3.1 MasterPersona タスクの `Completed` 遷移箇所を特定し、`task_id` 単位で queue クリーンアップを呼ぶ
- [ ] 3.2 `pkg/queue` または `llm_queue` 管理層に `Completed` 時のみ全 job を削除する処理を追加する
- [ ] 3.3 `Failed` / `Canceled` 時は job を保持する分岐を確認し、完了時削除と競合しないよう整理する

## 4. ドキュメントと検証

- [ ] 4.1 DB 変更がある場合は `openspec/specs/database_erd.md` の persona / queue 定義を更新する
- [ ] 4.2 ペルソナ一覧件数、生成リクエスト表示、画面再表示、Completed 後 queue 削除のテストケースを追加または更新する
- [ ] 4.3 `wails` または関連テストで MasterPersona の主要導線を検証し、結果を変更に記録する
