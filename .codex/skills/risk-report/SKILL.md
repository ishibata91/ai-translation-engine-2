---
name: risk-report
description: AI Translation Engine 2 専用。impl / fix flow の終端で `git diff` だけを材料にリスクを棚卸しし、Markdown レポートを生成する。
---

# Risk Report

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `risk-report` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は実装や修正の最終段で、差分リスクを Markdown へ要約する専用 skill。
コード変更、追加実装、再レビューは行わず、`git diff` に現れる事実だけを根拠に report を作成する。

## 使う場面
- `impl-direction` または `fix-direction` が review を通過した直後
- PR 送付前に差分起因リスクを可視化したい
- `residual risk` を会話ログではなく成果物(MD)として残したい

## 入力
- `invoked_skill: risk-report`
- `invoked_by: impl-direction | fix-direction`
- `change_id`
- `lane: impl | fix`
- `diff_range` (`working_tree` / `HEAD~1..HEAD` など)
- `focus` (既知の懸念領域があれば)

## 手順
1. `git diff --name-status <diff_range>` で変更ファイルを列挙する。
2. `git diff <diff_range>` で全文差分を確認する。
3. 変更を `仕様逸脱` `回帰可能性` `テスト不足` `運用影響` の4観点で棚卸しする。
4. 根拠が diff 上に無い推測は書かない。推測を置く場合は `仮説` と明記する。
5. `references/templates.md` の `Risk Report` フォーマットで Markdown を作る。
6. 出力先は `changes/<change_id>/context_board/<lane>-risk-report.md` とする。
7. 必要なら同名 `.summary.json` を補助出力してもよいが、正本は Markdown とする。

## 許可される動作
- `git diff` と `git status --short` の読解
- 既存 review feedback と突き合わせたリスク表現の正規化
- Markdown レポート生成

## 禁止
- コード修正
- テスト追加実行を前提にした評価の確定
- diff に無い設計議論の持ち込み

## 完了条件
- `changes/<change_id>/context_board/<lane>-risk-report.md` が作成済み
- レポートに `Overall` `Risk Items` `Open Questions` `Recommended Follow-up` が含まれる
- 各リスク項目に `diff根拠` が 1 つ以上書かれている
