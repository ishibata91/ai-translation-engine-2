---
name: aite2-implementation-review-guard
description: AI Translation Engine 2 専用。実装差分をレビューする。implementation skill から呼ばれる実装レビュー用 skill。
---

# AITE2 Implementation Review Guard

この skill は実装差分、change 文書、`docs/` 正本を照合し、重大な欠陥から順に指摘するための review skill。

## 使う場面
- backend / frontend / bug-fix 実装後の自己レビューをしたい
- implementation skill からレビュー修正ループを回したい
- PR 相当の差分レビューをしたい

## 手順
1. 関連する `changes/` 文書、`docs/` 正本、実装差分を集める。
2. バグ、退行、責務逸脱、仕様未反映、未同期、テスト不足の観点で見る。
3. Findings を重大度順に並べる。
4. 根拠となる文書または差分位置を結び付ける。
5. Open Questions と Residual Risks を分けて整理する。
6. docs 正本へ反映すべき仕様差分があるかを明示する。

## 呼び出し元 implementation skill との分担
- この skill 自身はコード修正を行わず、finding を返す
- 呼び出し元 implementation skill が finding を受けて修正する
- 修正後に再度この skill を呼び、前回 finding の解消有無を優先確認する
- 重大 / 中程度 finding がなくなるまでループする
- 同一論点が 2 周以上解消しない場合はユーザーへエスカレーションする

## 出力形式
- Findings を重大度順に並べる
- Open Questions を分ける
- Residual Risks を分ける
- change 修正要否と docs 同期要否を最後に短く添える

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に全論点を混ぜず、重大 finding から順に扱う
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- 実装方針の好みより仕様逸脱と退行を優先する
