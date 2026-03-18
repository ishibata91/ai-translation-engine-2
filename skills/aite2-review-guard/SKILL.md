---
name: aite2-review-guard
description: AI Translation Engine 2 専用。change 文書、仕様差分、実装差分を突き合わせ、バグ、責務逸脱、仕様未反映、仕様差分未同期、テスト不足をレビューする。レビュー依頼や自己点検時に使う。
---

# AITE2 Review Guard

この skill はレビュー用。
目的は、設計文書と実装のズレを種類ごとに切り分け、重大な欠陥から優先して指摘すること。

## 使う場面
- 実装後の自己レビュー
- PR 相当の差分レビュー
- 設計はあるが実装がそれに沿っているか不安
- UI / シナリオ / ロジックの取りこぼしを見たい

## 入力
- `docs/` の関連仕様
- 実装差分
- `changes/<id>/ui.md`
- `changes/<id>/scenarios.md`
- `changes/<id>/logic.md`
- 必要なら `tasks.md`

## 出力
- Findings
- Open Questions
- Residual Risks

## 観点
- バグや退行
- 責務境界違反
- 仕様差分の未確認
- `changes/` と `docs/` の未同期
- UI Contract 未反映
- シナリオ抜け
- テスト不足
- 品質ゲート不足

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に全論点を混ぜず、重大 finding から順に扱う
- 各ステップで確認済み範囲、未確認範囲、次に見る 1 論点を明確にする
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- 実装方針の好みより仕様逸脱と退行を優先する
- change 文書だけでなく `docs/` 正本との仕様差分も確認する
