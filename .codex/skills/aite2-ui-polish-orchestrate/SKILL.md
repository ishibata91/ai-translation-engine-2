---
name: aite2-ui-polish-orchestrate
description: AI Translation Engine 2 専用。既存 UI の見た目を整える。「余白を直して」「レイアウト崩れを直して」と言われたときに起動する。
---

# AITE2 UI Polish Orchestrate

この skill は ui-refine 系作業の指揮者として動き、既存 UI の見た目だけを対象に、観測、修正委譲、review、board handoff の順でデザイン差分を直すための orchestration skill。

## 使う場面
- 「余白を直して」と既存 UI の見た目修正を依頼された
- 「配置が崩れている」「文字が切れている」と既存画面の視認性改善を依頼された
- ロジックはそのままに、レイアウト、余白、配置、視認性だけを直したい

## 必読 spec
- `docs/frontend/ui-rules/spec.md`
- 補助: `docs/frontend/frontend-coding-standards/spec.md`

## handoff 前提
- ui-polish 指揮役は、原則として必要な前段 skill を subagent として起動する
- change が無い場合は `scripts/init-change-ui-refine-docs.ps1` で `changes/<id>/context_board/` を作る
- `plan-distill` が初期観測を整理した board を読む
- 修正方針と変更結果は board に残して次の skill へ渡す
- ロジック変更は board に明示的な指示が無い限り扱わない

## subagent 起動規約
- 初期観測や対象整理が必要な場合は `plan-distill` を使う
- 実際の UI 修正は `impl-frontend-work` を使う
- UI 実装差分の review が必要な場合は `impl-review` を使う

## 手順
1. `docs/frontend/ui-rules/spec.md` を読み、UI 生成ルールとレイアウト制約を確認する。
2. 必要なら `docs/frontend/frontend-coding-standards/spec.md` で実装上の制約を確認する。
3. board から対象画面、対象要素、対象ファイルを特定する。
4. 現状観測を言語化し、見た目の問題を board に追記する。
5. 余白、配置、視認性、整列のどこを直すかを最小単位で決める。
6. `impl-frontend-work` に、ロジック変更を混ぜない UI 修正として実装を委譲する。
7. 必要なら `impl-review` に review を委譲する。
8. 修正前後の差分と残リスクを board に残す。

## 参照資料
- 起動例と非起動例は `references/examples.md` を読む。
- 観測メモと修正メモは `references/templates.md` を使う。
- 修正前の確認項目は `references/checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に複数箇所へ広げず、指定された対象から順に直す
- 各ステップで対象、観測結果、次の 1 手を明確にする
- 指揮役は orchestration 以外を行わず、自分で修正や review をしない
- 指定されていないファイルへ勝手に範囲を広げない
- デザイン修正にロジック変更を混ぜない
- board を更新せずに次の skill へ handoff しない
