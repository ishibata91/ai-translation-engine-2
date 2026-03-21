---
name: plan-review
description: AI Translation Engine 2 専用。設計差分と docs 正本をレビューする。差分仕様、UI、シナリオ、ロジックの整合性を implementation 前に確認したいときに使う。
---

# Plan Review

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `plan-review` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は設計差分と docs 正本の整合をレビューする skill。
change 文書、`docs/` 正本、architecture を照合し、設計上の重大な欠陥から順に指摘する。

## 使う場面
- 実装前に設計レビューをしたい
- `logic.md` / `ui.md` / `scenarios.md` と `docs/` のズレを見たい
- docs 正本へ上げるべき仕様断片を確認したい

## 入力契約
- `plan-direction` から渡された review 対象
- `plan-direction`以外から呼び出された場合､生成を拒否して`plan-direction`へ親を返すこと｡
- 関連する `changes/` 文書
- 参照すべき `docs/` 正本
- review で重点確認したい論点

## 手順
1. 関連する `changes/` 文書と `docs/` 正本を集める。
2. 以下のアーキテクチャ定義およびコーディング規約に従っているかチェックする。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. 責務逸脱、仕様未反映、未同期、シナリオ抜けの観点で見る。
4. Findings を重大度順に並べる。
5. 根拠となる文書位置を結び付ける。
6. Open Questions と Residual Risks を分けて整理する。
7. `plan-sync` に渡すべき昇格候補があるかを明示する。

## 出力形式
- `score` (0.0 - 1.0)
- Findings を重大度順に並べる
- Open Questions を分ける
- Residual Risks を分ける
- docs 同期要否を最後に短く添える

## 参照資料
- findings 記録は `references/templates.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 実装差分の是正ループには入らない
- まず findings を重大度順で出す
- 要約は findings の後に置く
- 問題なしならそれを明示する
- change 文書だけでなく `docs/` 正本との仕様差分も確認する
- read-only で扱い、実装差分の是正ループには入らない
- `score >= 0.85` を満たした review だけを次工程へ渡してよい
