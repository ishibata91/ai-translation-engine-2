## 1. Workflow と terminology 契約整理

- [x] 1.1 translation flow workflow に `単語翻訳` phase を追加し、`データロード -> 単語翻訳 -> 本文翻訳` の順序を固定する
- [x] 1.2 workflow から terminology slice へ task 識別子と terminology 用 request/prompt 設定 DTO だけを渡す契約を追加する
- [x] 1.3 terminology slice 側で artifact から対象レコードを抽出し、`PreparePrompts` と `SaveResults` に必要な入力組み立てを完結させる

## 2. terminology 永続化と状態管理

- [x] 2.1 terminology の保存先を単一 `terminology.db` に切り替え、ファイル名ベースの mod テーブル解決を実装する
- [x] 2.2 terminology 側に phase 状態用 `status` カラムとサマリ取得処理を追加し、`対象件数 / 保存件数 / 失敗件数` を返せるようにする
- [ ] 2.3 `terminology.db` の構造変更に合わせて関連 ERD / spec を更新する

## 3. shared frontend component 拡張

- [x] 3.1 `frontend/src/components/ModelSettings.tsx` を terminology phase でも再利用できるようにし、namespace と表示文言を props で切り替えられるようにする
- [x] 3.2 `frontend/src/components/masterPersona/PromptSettingCard.tsx` を terminology phase でも再利用できるようにし、補助文言とカード情報を props で差し替えられるようにする
- [x] 3.3 terminology 用の request 設定と prompt 設定を frontend 側 config namespace に保存・復元する hook / 型を追加する

## 4. Translation Flow UI 統合

- [x] 4.1 `frontend/src/components/translation-flow/TerminologyPanel.tsx` を実装し、単語翻訳 phase の状態表示、実行サマリ、進行表示を実データに接続する
- [x] 4.2 TerminologyPanel に `ModelSettings` と `PromptSettingCard` を組み込み、terminology 用 request/prompt 設定を編集できるようにする
- [x] 4.3 translation flow 再訪時に terminology の phase 状態、実行サマリ、保存済み設定を復元表示できるようにする

## 5. 検証と品質ゲート

- [x] 5.1 backend 変更ファイルに対して `npm run backend:lint:file -- <file...>` を順次実行し、最後に `npm run lint:backend` を通す
- [x] 5.2 frontend 変更ファイルに対して `npm run lint:file -- <file...>` を順次実行し、最後に `npm run typecheck` と `npm run lint:frontend` を通す
- [x] 5.3 translation flow と shared settings UI の Playwright / 必要テストを更新し、単語翻訳 phase と request/prompt 編集の必須シナリオを検証する
