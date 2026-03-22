# Implementation Routing Examples

## frontend へ送る例
- `ui.md に沿って画面を実装して`
- `この画面の state を React に反映して`
- `frontend の task を進めて`

## backend へ送る例
- `logic.md に沿って workflow を実装して`
- `slice と artifact の境界をコードに反映して`
- `backend の task を進めて`

## mixed（分割して扱う）例
- `画面と API の両方を実装して`
- distill 返却後、`impl-workplan` でモジュール/契約単位の section plan を作る
- shared contract（型・API 形式）を先に固定し、その後 section ごとに worker を起動する

## `ui.md` 不在でも frontend を含む例
- 依頼: `既存画面の状態遷移と backend API を一緒に更新して`
- 挙動: `ui.md` の有無だけで backend-only にしない
- 判定: `frontend/src` 変更や frontend 品質ゲートが section signal に含まれるなら frontend section を残す

## review 差し戻し例
- `impl-review` が `affected_sections` に `ui-state-sync` の full section contract と `required_delta` を返す
- `impl-direction` は `Review Reroute` を作り、`shared_contract` `required_reading` `validation_commands` `acceptance` を落とさず `ui-state-sync` だけ再 dispatch する

## plan 依頼の誤投入
- 依頼: `新しい設定画面の仕様を決めたい`
- 挙動: `impl-direction` では処理を進めず停止する
- handoff 先: `plan-direction`

## bugfix 依頼の誤投入
- 依頼: `再開後に進捗表示が消えた原因を調べて`
- 挙動: `impl-direction` では処理を進めず停止する
- handoff 先: `fix-direction`
