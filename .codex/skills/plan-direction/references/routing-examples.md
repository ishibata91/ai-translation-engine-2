# Design Routing Examples

## 既存見た目修正
- 依頼: `この画面の余白を直して`
- 挙動: `plan-direction` では処理を進めず停止する
- handoff 先: `impl-direction`

## UI を含む新規変更
- 依頼: `新しい設定画面の仕様を決めたい`
- 最初の skill: `plan-direction`
- 自律チェーン: `plan-distill` -> `plan-ui` -> `plan-scenario` -> `plan-logic` -> `plan-review`

## UI を含まない仕様変更
- 依頼: `再試行フローのシナリオと内部責務を整理したい`
- 最初の skill: `plan-direction`
- 自律チェーン: `plan-distill` -> `plan-scenario` -> `plan-logic` -> `plan-review`

## docs 同期
- 依頼: `change 文書を docs に反映したい`
- 最初の skill: `plan-direction`
- 自律チェーン: `plan-distill` -> `plan-sync`

## 実装修正の誤投入
- 依頼: `TerminologyPanel.tsx を直したい`
- 挙動: `plan-direction` では処理を進めず停止する
- handoff 先: `impl-direction`

## bugfix の誤投入
- 依頼: `再開後に進捗バーが消える不具合を調査して`
- 挙動: `plan-direction` では処理を進めず停止する
- handoff 先: `fix-direction`
