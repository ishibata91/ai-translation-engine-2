---
name: aite2-implementation-review-guard
description: AI Translation Engine 2 専用。実装差分をレビューする。implementation skill から呼ばれる実装レビュー用 skill。
---

# AITE2 Implementation Review Guard

この skill は reviewer 向けの read-only 手順を定義し、実装差分、change 文書、`docs/` 正本を照合して重大な欠陥から順に指摘するための review skill。

## 使う場面
- backend / frontend / bug-fix 実装後の自己レビューをしたい
- implementation skill からレビュー修正ループを回したい
- PR 相当の差分レビューをしたい

## reviewer 入力契約
- 対象 change
- 関連する `changes/` 文書
- 参照すべき `docs/` 正本
- 実装差分
- 実行済み検証結果
- 前回 findings

## subagent 実行前提
- この skill は bugfix / implementation の指揮者から subagent として起動される前提で使う
- subagent 起動時は `agents/openai.yaml` の profile 設定を使う

## 手順
1. 関連する `changes/` 文書、`docs/` 正本、実装差分、検証結果、前回 findings を集める。
2. バグ、退行、責務逸脱、仕様未反映、未同期、テスト不足の観点で見る。
3. Findings を `critical` / `medium` / `low` の重大度順に並べる。
4. 根拠となる文書または差分位置を結び付ける。
5. Open Questions と Residual Risks を分けて整理する。
6. docs 正本へ反映すべき仕様差分があるかを明示する。

## 呼び出し元 implementation skill との分担
- この skill 自身はコード修正を行わず、finding を返す
- 呼び出し元 implementation skill が finding を受けて修正する
- 修正後に再度この skill を呼び、前回 finding の解消有無を優先確認する
- `critical` / `medium` finding がなくなるまでループする
- 同一論点が 2 周以上解消しない場合はユーザーへエスカレーションする

## 出力形式
- Findings を重大度順に並べる
- Open Questions を分ける
- Residual Risks を分ける
- change 修正要否と docs 同期要否を最後に短く添える

## 終了条件
- `critical` と `medium` finding が 0 件なら review loop を終了してよい
- `low` のみ残る場合は Residual Risks として明示して返す
- 問題なしなら Findings なしを明示する

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に全論点を混ぜず、重大 finding から順に扱う
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- 実装方針の好みより仕様逸脱と退行を優先する
- reviewer は read-only として振る舞う
