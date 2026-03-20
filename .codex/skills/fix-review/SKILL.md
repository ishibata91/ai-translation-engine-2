---
name: fix-review
description: AI Translation Engine 2 専用。bugfix 差分をレビューし、退行、未解消リスク、docs handoff 要否を返す read-only reviewer。修正後の bugfix review をしたいときに使う。
---

# Fix Review

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-review` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は bugfix 差分をレビューし、未解消の退行や仕様逸脱を返す read-only skill。
bugfix 差分、関連仕様、検証結果を照合して結果を返す。

## 使う場面
- bugfix 修正後の退行確認をしたい
- 修正で別の contract を壊していないか見たい
- docs 反映が必要かを判定したい

## 入力契約
- `fix-direction` から渡された対象 change
- bugfix packet または fix plan
- 実装差分
- 実行済み検証結果
- 前回 findings

## 手順
1. bugfix scope と実装差分を照合する。
2. 退行、未解消リスク、仕様逸脱、未検証を優先して見る。
3. `severity` `location` `violated_contract` `required_delta` `recheck` を返す。
4. docs 反映が必要な場合だけ `docs_sync_needed` を示す。
5. 次工程の起動は行わず、結果だけを `fix-direction` へ返す。

## 出力形式
- `score` (0.0 - 1.0)
- `severity`
- `location`
- `violated_contract`
- `required_delta`
- `recheck`
- `docs_sync_needed`

## `fix-direction` が判断する条件
- `score >= 0.85` の review だけを次工程へ渡してよい
- `score < 0.85` の場合は `required_delta` を返して review loop を継続する

## 原則
- 実装方針の好みより退行と未解消リスクを優先する
- read-only として振る舞う
- review 結果は返すが、`fix-work` や `plan-sync` など次工程を自分で起動しない
