---
name: fix-review
description: AI Translation Engine 2 専用。bugfix 差分をレビューし、退行、未解消リスク、docs handoff 要否を返す。修正後の bugfix review をしたいときに使う。
---

# Fix Review

この skill は `review_cycler` 用の bugfix review skill。
bugfix 差分、関連仕様、検証結果を照合し、未解消の退行や仕様逸脱を返す。

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
- `severity`
- `location`
- `violated_contract`
- `required_delta`
- `recheck`
- `docs_sync_needed`

## 原則
- 実装方針の好みより退行と未解消リスクを優先する
- `review_cycler` は read-only として振る舞う
