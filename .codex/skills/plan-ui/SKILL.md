---
name: plan-ui
description: AI Translation Engine 2 専用。UI 契約と観測可能な振る舞いを設計する。画面仕様、state、導線、UI 事実を差分仕様として確定したいときに使う。
---

# Plan UI

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `plan-ui` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は UI 契約と観測可能な振る舞いを `ui.md` に落とす skill。
画面責務、主要 state、導線、観測可能な UI 事実を整理して返す。

## 使う場面
- 新しい画面やダイアログの UI 契約を決めたい
- 既存 UI の state や導線が曖昧で整理したい
- 実装前に state machine と観測可能な UI 事実を固めたい
- 実装前に UI 契約を change 文書へ落としたい

## 必読 spec
- `docs/frontend/ui-rules/spec.md`

## 入力契約
- `plan-direction` から渡された planning packet
- 対象 change
- 対象画面または導線
- `changes/<id>/ui.md` に落とすための決定事項

## 手順
1. `docs/frontend/ui-rules/spec.md` を読み、UI 生成ルールとレイアウト制約を確認する。
2. context packet から対象画面の Purpose と Primary Action を確定する。
3. 主要 state を洗い出し、Mermaid の state machine に落とす。
4. 各 state で観測可能な UI 事実を定義する。
5. 成功、失敗、空、待機、再試行の扱いを整理する。
6. 画面構造と確認観点をまとめ、`changes/<id>/ui.md` の形へ落とす。

## 参照資料
- UI 文書の雛形は `references/templates.md` を使う。
- 代表的な state machine の例は `references/examples.md` を読む。
- state の詰め方は `references/state-checklist.md` を使う。

## 許可される動作
- UI 契約の確定に集中し、実装判断は下流 lane の判断材料へ委ねる
- state は Mermaid の state machine で定義する
- 各 state で観測可能な UI 事実を書く
- 生成・更新対象は `ui.md` に限る
- 未確定事項は `Open Questions` に残す
