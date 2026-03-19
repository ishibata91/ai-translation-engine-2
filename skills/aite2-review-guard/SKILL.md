---
name: aite2-review-guard
description: AI Translation Engine 2 専用。change と実装をレビューする。「レビューして」「仕様と実装のズレを見て」と言われたときに起動する。
---

# AITE2 Review Guard

この skill は change 文書、`docs/` 正本、実装差分を照合し、重大な欠陥から順に指摘するための review skill。

## 使う場面
- 実装後の自己レビューをしたい
- PR 相当の差分レビューをしたい
- 設計と実装が揃っているか確認したい
- UI / シナリオ / ロジックの取りこぼしを見たい

## 手順
1. 関連する `changes/` 文書、`docs/` 正本、実装差分を集める。
2. バグ、退行、責務逸脱、仕様未反映、未同期、テスト不足の観点で見る。
3. Findings を重大度順に並べる。
4. 根拠となる文書または差分位置を結び付ける。
5. Open Questions と Residual Risks を分けて整理する。

## 参照資料
- レビュー記録の雛形は `references/templates.md` を使う。
- 指摘の例は `references/examples.md` を読む。
- 観点漏れ防止は `references/review-checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に全論点を混ぜず、重大 finding から順に扱う
- 各ステップで確認済み範囲、未確認範囲、次に見る 1 論点を明確にする
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- 実装方針の好みより仕様逸脱と退行を優先する
- change 文書だけでなく `docs/` 正本との仕様差分も確認する
