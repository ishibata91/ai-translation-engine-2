# Tasks

## 1. Workflow Contract
- [x] `pkg/workflow` の terminology preview DTO に `translated_text` と `translation_state` を追加する
- [x] `pkg/workflow` の terminology phase result DTO に `status` `progress_mode` `progress_current` `progress_total` `progress_message` を追加する
- [x] controller binding と Wails payload で上記 contract を透過させる

## 2. Terminology Read Model
- [x] terminology slice で preview row ID を key にした translated text lookup を返せるようにする
- [x] duplicate 統合と NPC FULL/SHRT ペア翻訳結果を preview row ID へ fan-out する read model を定義する
- [x] `completed_partial` と `run_error` を返せる phase summary 契約へ整理する

## 3. Progress Publication
- [x] workflow が runtime 進捗を `translation_flow.terminology.progress` へ流せるようにする
- [x] `GetTranslationFlowTerminology` で最新 progress snapshot を再取得できるようにする
- [x] 実行中に `progress_mode=determinate|indeterminate|hidden` を一貫して返す

## 4. Frontend Terminology Phase
- [x] terminology target table に `Translated Text` 列を追加し、`translation_state=missing` を `未翻訳` と表示する
- [x] 実行中は対象単語リスト本文を loading 表示に切り替え、`状態を再読込`、pagination、`単語翻訳を確定して次へ` を無効化する
- [x] `単語翻訳を実行` ボタン横に inline progress cluster を表示し、`progress_mode` に応じて determinate / indeterminate を切り替える
- [x] 既存 task 再表示時に translated text、summary、progress snapshot を復元する

## 5. Verification
- [x] workflow / slice の contract テストで translated text lookup、status 遷移、progress snapshot を確認する
- [x] frontend の adapter / state テストで DTO 拡張と UI state mapping を確認する
- [x] E2E で `Translated Text` 列、実行中 progress、一覧 loading、partial completion、empty/error、task 再表示の復元を確認する
