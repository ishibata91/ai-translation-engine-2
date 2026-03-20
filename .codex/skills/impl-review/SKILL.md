---
name: impl-review
description: AI Translation Engine 2 専用。実装差分をレビューし、required delta を返す。Observation Masking 前提で統合差分を検査したいときに使う。
---

# Impl Review

この skill は `review_cycler` 用の実装レビュー skill。
spec 抜粋、統合差分、検証結果、前回 findings を照合し、required delta 中心で返す。

## 使う場面
- 実装後の統合差分を read-only でレビューしたい
- 退行、contract 逸脱、未検証を優先確認したい
- docs 反映が必要かを判定したい

## 入力契約
- 対象 change
- `spec_excerpt`
- `structured_diff`
- `verification`
- `previous_findings`

## 手順
1. `spec_excerpt` と `structured_diff` を照合する。
2. 退行、contract 逸脱、責務逸脱、未検証を優先して見る。
3. `severity` `location` `violated_contract` `required_delta` `recheck` を返す。
4. docs 反映が必要な仕様差分だけ `docs_sync_needed` に示す。

## 出力形式
- `severity`
- `location`
- `violated_contract`
- `required_delta`
- `recheck`
- `docs_sync_needed`

## 終了条件
- `critical` と `medium` が 0 件なら review loop を終了してよい
- `low` のみ残る場合は残留リスクとして返す

## 原則
- Observation Masking 前提で、不要な背景説明を受け取らない
- 実装方針の好みより仕様逸脱と退行を優先する
- `review_cycler` は read-only として振る舞う
