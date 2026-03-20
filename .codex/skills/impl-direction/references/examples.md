# Implementation Routing Examples

## frontend へ送る例
- `ui.md に沿って画面を実装して`
- `この画面の state を React に反映して`
- `frontend の task を進めて`

## backend へ送る例
- `logic.md に沿って workflow を実装して`
- `slice と artifact の境界をコードに反映して`
- `backend の task を進めて`

## 分割して扱う例
- `画面と API の両方を実装して`
- まず frontend task と backend task に分ける

## plan 依頼の誤投入
- 依頼: `新しい設定画面の仕様を決めたい`
- 挙動: `impl-direction` では処理を進めず停止する
- handoff 先: `plan-direction`

## bugfix 依頼の誤投入
- 依頼: `再開後に進捗表示が消えた原因を調べて`
- 挙動: `impl-direction` では処理を進めず停止する
- handoff 先: `fix-direction`
