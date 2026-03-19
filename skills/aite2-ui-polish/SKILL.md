---
name: aite2-ui-polish
description: AI Translation Engine 2 専用。既存 UI の見た目を整える。「余白を直して」「レイアウト崩れを直して」と言われたときに起動する。
---

# AITE2 UI Polish

この skill は既存 UI の見た目だけを対象に、観測、最小修正、実画面確認の順でデザイン差分を直すための skill。

## 使う場面
- 「余白を直して」と既存 UI の見た目修正を依頼された
- 「配置が崩れている」「文字が切れている」と既存画面の視認性改善を依頼された
- ロジックはそのままに、レイアウト、余白、配置、視認性だけを直したい

## 必読 spec
- `docs/frontend/ui-rules/spec.md`
- 補助: `docs/frontend/frontend-coding-standards/spec.md`

## 手順
1. `docs/frontend/ui-rules/spec.md` を読み、UI 生成ルールとレイアウト制約を確認する。
2. 必要なら `docs/frontend/frontend-coding-standards/spec.md` で実装上の制約を確認する。
3. 対象画面と対象ファイルを特定する。
4. `wails dev` と Playwright MCP で現状を観測し、見た目の問題を言語化する。
5. 余白、配置、視認性、整列のどこを直すかを最小単位で決める。
6. ロジック変更を混ぜず、対象範囲だけを最小差分で修正する。
7. 同じ画面を Playwright MCP で再確認し、修正前後の差分を確認する。
8. フロントの品質ゲートを通す。

## 参照資料
- 起動例と非起動例は `references/examples.md` を読む。
- 観測メモと修正メモは `references/templates.md` を使う。
- 修正前の確認項目は `references/checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に複数箇所へ広げず、指定された対象から順に直す
- 各ステップで対象、観測結果、次の 1 手を明確にする
- 指定されていないファイルへ勝手に範囲を広げない
- デザイン修正にロジック変更を混ぜない
- `wails dev` の起動方法を勝手に変えない
- Playwright MCP で実画面確認するまで完了扱いにしない
