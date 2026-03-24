---
name: impl-review
description: AI Translation Engine 2 専用。実装差分をレビューし、required delta を返す。修正や差分編集は行わず、Observation Masking 前提で統合差分を検査したいときに使う。
---

# Impl Review

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-review` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は実装差分をレビューし、required delta を返すだけの skill。
spec 抜粋、統合差分、検証結果、前回 findings を照合し、required delta 中心で返す。

## 使う場面
- 実装後の統合差分を read-only でレビューしたい
- 退行、contract 逸脱、未検証を優先確認したい
- docs 反映が必要かを判定したい

## 入力契約
- 対象 change
- `tasks.md` または section map
- `spec_excerpt`
- `structured_diff`
- `verification`
  - frontend 差分が含まれる場合は `npm --prefix frontend run build` の結果を含める
- `previous_findings`

## 手順
1. `spec_excerpt` と `structured_diff` を照合する。
2. 以下の規約およびアーキテクチャ定義に従っているかチェックする。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. 退行、contract 逸脱、責務逸脱、未検証を優先して見る。frontend 差分が含まれるのに `npm --prefix frontend run build` の結果が無い、または失敗している場合は required delta または residual risk として見落とさない。
4. `severity` `location` `affected_sections` `violated_contract` `required_delta` `recheck` を返す。
5. docs 反映が必要な仕様差分だけ `docs_sync_needed` に示す。
6. 自分では修正、差分編集、worker への差し戻し実行を行わず、レビュー結果だけを返して終了する。

## 出力形式
- 正本は `changes/<id>/context_board/impl-review.feedback.json` とし、field は `references/templates.md` に完全一致させる
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、`impl-review.feedback.validation.json` を出力する
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する

## 終了条件
- `score >= 0.85` かつ `critical` と `medium` が 0 件で、frontend 差分がある場合は `npm --prefix frontend run build` の結果が verification に含まれているなら review loop を終了してよい
- `score < 0.85` の場合は `required_delta` を返して review loop を継続する
- `low` のみ残る場合は残留リスクとして返す

## 原則
- Observation Masking 前提で、不要な背景説明を受け取らない
- 実装方針の好みより仕様逸脱と退行を優先する
- frontend 差分で build 結果が無いレビューを green 扱いにしない
- read-only として振る舞う
- 自分では修正、差分編集、worker 起動、差し戻し実行を行わず、レビュー結果だけを返す
