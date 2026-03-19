---
name: aite2-design-review-guard
description: AI Translation Engine 2 専用。change と docs 正本をレビューする。実装前の設計レビューを行いたいときに起動する。
---

# AITE2 Design Review Guard

この skill は change 文書、`docs/` 正本、architecture を照合し、設計上の重大な欠陥から順に指摘するための review skill。

## 使う場面
- 実装前に設計レビューをしたい
- `logic.md` / `ui.md` / `scenarios.md` と `docs/` のズレを見たい
- docs 正本へ上げるべき仕様断片を確認したい

## 手順
1. 関連する `changes/` 文書と `docs/` 正本を集める。
2. 必要に応じて `docs/governance/architecture/spec.md` を確認する。
3. 責務逸脱、仕様未反映、未同期、シナリオ抜けの観点で見る。
4. Findings を重大度順に並べる。
5. 根拠となる文書位置を結び付ける。
6. Open Questions と Residual Risks を分けて整理する。
7. `aite2-sync-docs` に渡すべき昇格候補があるかを明示する。

## 出力形式
- Findings を重大度順に並べる
- Open Questions を分ける
- Residual Risks を分ける
- docs 同期要否を最後に短く添える

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 実装差分の是正ループには入らない
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- change 文書だけでなく `docs/` 正本との仕様差分も確認する
