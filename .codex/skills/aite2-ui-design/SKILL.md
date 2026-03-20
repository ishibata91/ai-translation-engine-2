---
name: aite2-ui-design
description: AI Translation Engine 2 専用。UI 仕様を設計する。「画面仕様を決めたい」「UI 契約を整理したい」と言われたときに起動する。
---

# AITE2 UI Design

この skill は architect 向けに、UI 実装前の画面責務、主要 state、導線、検証観点を整理するための design skill。

## 使う場面
- 新しい画面やダイアログの UI 契約を決めたい
- 既存 UI の state や導線が曖昧で整理したい
- 実装前に state machine と観測可能な UI 事実を固めたい
- 実装前に UI 契約を change 文書へ落としたい

## 必読 spec
- `docs/frontend/ui-rules/spec.md`

## architect 入力契約
- `aite2-design-orchestrate` から渡された context packet
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

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に複数の設計判断を確定しようとしない
- 各ステップで決めたこと、未確定事項、次の 1 手を明確にする
- UI はロジック詳細を持たない
- 状態を先に固定し、コンポーネント分割は後に考える
- 画面ごとの差より共通パターンを優先する
- 未確定事項は曖昧なまま実装へ流さない
- state は Mermaid の state machine で定義する
- state 名だけでなく、各 state で観測可能な UI 事実を書く
- architect は UI 契約の確定に集中し、実装確認ループへ入らない
