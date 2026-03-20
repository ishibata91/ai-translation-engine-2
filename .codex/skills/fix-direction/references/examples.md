# Bug Fix Examples

## 起動例
- `翻訳開始ボタンが反応しないのを修正して`
- `バグ: 保存後に一覧が更新されない`
- `バグ: 再開後に進捗表示が消えたままになる`
- `この不具合、docs と実装がズレていないか見て`

## 起動しない例
- `新しい翻訳設定画面を設計して`
- `Master Persona のシナリオを整理して`
- `docs/ のタイトルを見直して`

## plan 依頼の誤投入
- 依頼: `新しい翻訳設定画面の仕様を決めたい`
- 挙動: `fix-direction` では処理を進めず停止する
- handoff 先: `plan-direction`

## impl 依頼の誤投入
- 依頼: `TerminologyPanel.tsx の state を React に反映して`
- 挙動: `fix-direction` では処理を進めず停止する
- handoff 先: `impl-direction`

## UI バグの例
依頼:
`バグ: 実行中に Cancel を押してもボタンが disabled のままで戻らない`

期待される進め方:
- `fix-direction` 自身は対象画面の特定、spec 読解、再現、ログ観測を直接行わない
- `fix-distill` に再現条件、関連仕様、関連コード、既知観測の整理を委譲する
- `fix-trace` に原因仮説と観測計画の作成を委譲する
- 再現後は `fix-analysis` にログ整理を委譲する
- fix scope 確定後に `fix-work` と `fix-review` を順に起動する

## Backend バグの例
依頼:
`辞書作成が途中で止まるのを修正して`

期待される進め方:
- `fix-direction` 自身は再現条件固定、spec 照合、ログ確認を直接行わない
- `fix-distill` と `fix-trace` に停止地点の切り分け材料を集めさせる
- 再現後は `fix-analysis` に事実整理を委譲する
- fix scope 確定後に `fix-work` へ最小修正を渡し、最後に `fix-review` を起動する

## 仕様乖離の例
依頼:
`バグ: 翻訳完了後に結果パネルが自動で開かない`

切り分け観点:
- `docs/` に自動展開が明記されているなら実装不備
- 現行要件では自動展開しないなら、報告内容ではなく spec を優先する
- spec が曖昧ならユーザー確認を挟む
