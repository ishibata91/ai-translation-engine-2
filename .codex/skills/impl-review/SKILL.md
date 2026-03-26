---
name: impl-review
description: AI Translation Engine 2 専用。実装差分をレビューし、required delta を返す。修正や差分編集は行わず、Observation Masking 前提で統合差分を検査したいときに使う。
---

# Impl Review

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-review` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は実装差分をレビューし、required delta を返すだけの skill。
spec 抜粋、統合差分、検証結果、前回 findings を照合し、required delta 中心で返す。

## 必須レビュー観点
- 仕様適合レビュー: `spec_excerpt` と差分を照合し、仕様未達、contract 逸脱、受け入れ条件抜けを確認する
- 差分の危険箇所レビュー: 変更の波及範囲、既存動作への退行リスク、責務境界の破壊を確認する
- 原則適用レビュー: `design_principles`（SRP/SoC/DIP/OCP）への適合と逸脱理由の妥当性を確認する
- テスト不足レビュー: 変更に対して必要な検証、build、typecheck、lint、再現手順が足りているかを確認する
- 例外・失敗時レビュー: エラー処理、失敗時の挙動、境界条件、未入力や不正入力時の扱いを確認する
- 既存設計との整合性レビュー: architecture、既存 contract、責務分離、既存パターンとの整合を確認する
- セキュリティ・性能レビュー: 権限、入力起点の危険、不要な高コスト処理、N+1 や無駄な再計算の持ち込みを確認する

## 使う場面
- 実装後の統合差分を read-only でレビューしたい
- 退行、contract 逸脱、未検証を優先確認したい
- docs 反映が必要かを判定したい

## 入力契約
- 対象 change
- `tasks.md` または section map
- `spec_excerpt`
- `structured_diff`
- `design_principles`（section ごとの work order で固定された原則）
- `verification`
  - frontend 差分が含まれる場合は `npm --prefix frontend run build` の結果を含める
- `previous_findings`

## 手順
1. `spec_excerpt` と `structured_diff` を照合する。
2. 以下の規約およびアーキテクチャ定義に従っているかチェックする。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. 必須レビュー観点を順に確認する。観点を飛ばさず、仕様適合、危険箇所、原則適用、テスト不足、例外・失敗時、既存設計整合性、セキュリティ・性能の順で見落としを潰す。
4. 退行、contract 逸脱、責務逸脱、未検証を優先して見る。frontend 差分が含まれるのに `npm --prefix frontend run build` の結果が無い、または失敗している場合は最低でも `medium` として扱い、`score` を `0.75` 以下にする。
5. `severity` は未解消の最高重大度を返す。`critical` `medium` `low` の 3 段階を使い、好みや任意改善は指摘理由に含めない。
6. `score` は「欠陥の重さ」を表す離散バンドで返す。判定優先順位は `critical > medium > verification不足 > low件数 > external noise` とし、複数条件がある場合は最も低いバンドを採用する。
7. `score` は以下の rubric に従う。
   - `1.00`: 未解消の品質欠陥なし。required verification も満たしている。
   - `0.90`: 未解消が `low` 1-2 件のみ、または `external_validation_noise` / `known_pre_existing_issue` のみ。
   - `0.85`: 未解消が `low` 3-4 件のみ。
   - `0.75`: 未解消の `medium` が 1 件以上ある、required verification 不足がある、または `low` が 5 件以上ある。
   - `0.50`: 未解消の `critical` が 1 件以上ある。
8. `score` `severity` `location` `affected_sections` `violated_contract` `required_delta` `recheck` を返す。件数や広がりは `required_delta` と `recheck` で説明し、`affected_sections` は reroute 用にだけ使う。
9. docs 反映が必要な仕様差分だけ `docs_sync_needed` に示す。`docs_sync_needed` は score に影響させない。
10. 自分では修正、差分編集、worker への差し戻し実行を行わず、レビュー結果だけを返して終了する。

## 出力形式
- 正本は `changes/<id>/context_board/impl-review.feedback.json` とし、field は `references/templates.md` に完全一致させる
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、`impl-review.feedback.validation.json` を出力する
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する

## 終了条件
- `score >= 0.85` かつ `critical` と `medium` が 0 件で、frontend 差分がある場合は `npm --prefix frontend run build` の結果が verification に含まれているなら review loop を終了してよい
- `score < 0.85` の場合は `required_delta` を返して review loop を継続する
- `low` のみ残る場合でも、3-4 件までは残留リスクとして返してよいが、5 件以上なら `score = 0.75` として review loop を継続する

## 許可される動作
- Observation Masking 前提で、不要な背景説明を受け取らない
- 実装方針の好みより仕様逸脱と退行を優先する
- green 判定は、必須レビュー観点を一通り確認した後に行う
- frontend 差分を green 扱いにするのは、build 結果が verification に含まれる場合に限る
- `score >= 0.85` とするのは、`critical` / `medium` が未解消でない場合に限る
- `external_validation_noise` と `known_pre_existing_issue` は通常欠陥より軽く扱うが、score 上限は `0.90` とする
- `docs_sync_needed` は score 減点理由ではなく handoff 判断材料として扱う
- read-only として振る舞う
- 返却内容はレビュー結果に限り、修正や worker 起動は `impl-direction` に委ねる
