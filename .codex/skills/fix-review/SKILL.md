---
name: fix-review
description: AI Translation Engine 2 専用。bugfix 差分をレビューし、退行、未解消リスク、docs handoff 要否を返す。修正後の bugfix review をしたいときに使う。
---

# Fix Review

この skill は bugfix 差分をレビューし、未解消の退行や仕様逸脱を返す skill。
bugfix 差分、関連仕様、検証結果を照合して結果を返す。

## 使う場面
- bugfix 修正後の退行確認をしたい
- 修正で別の contract を壊していないか見たい
- docs 反映が必要かを判定したい

## 入力契約
- 対象 change
- bugfix packet または fix plan
- 実装差分
- 実行済み検証結果
- 前回 findings

## 手順
1. bugfix scope と実装差分を照合する。
2. 退行、未解消リスク、仕様逸脱、未検証を優先して見る。
3. `severity` `location` `violated_contract` `required_delta` `recheck` を返す。
4. docs 反映が必要な場合だけ `docs_sync_needed` を示す。

## 出力形式
- `score` (0.0 - 1.0)
- `severity`
- `location`
- `violated_contract`
- `required_delta`
- `recheck`
- `docs_sync_needed`

## 次に進める条件
- `score >= 0.85` の review だけを次工程へ渡してよい
- `score < 0.85` の場合は `required_delta` を返して review loop を継続する

## 原則
- 実装方針の好みより退行と未解消リスクを優先する
- read-only として振る舞う
