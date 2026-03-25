---
name: fix-review
description: AI Translation Engine 2 専用。bugfix 差分をレビューし、退行、未解消リスク、docs handoff 要否を返す read-only reviewer。修正後の bugfix review をしたいときに使う。
---

# Fix Review

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-review` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は bugfix 差分をレビューし、未解消の退行や仕様逸脱を返す read-only skill。
bugfix 差分、関連仕様、検証結果を照合して結果を返す。

## 必須レビュー観点
- 仕様適合レビュー: bugfix scope、関連仕様、期待挙動に対して修正が過不足なく適合しているかを確認する
- 差分の危険箇所レビュー: 修正点の周辺、再発しやすい条件、関連フローへの退行リスクを確認する
- テスト不足レビュー: 再現手順、回帰確認、追加ケース、既存検証で不足しているものがないかを確認する
- 例外・失敗時レビュー: 修正後も失敗系、境界条件、エラー復旧、部分失敗時の挙動が破綻しないかを確認する
- 既存設計との整合性レビュー: 応急処置で責務境界や既存 contract を壊していないかを確認する
- セキュリティ・性能レビュー: 修正に伴う入力起点の危険、権限逸脱、過負荷、無駄なリトライや重複処理を確認する
- DB メンテ回避レビュー: DB メンテナンスで解消できる問題を、起動時 self-heal やデータ補正コードで覆い隠していないかを確認する

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
2. 以下の規約およびアーキテクチャ定義に従っているかチェックする。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. 必須レビュー観点を順に確認する。観点を飛ばさず、仕様適合、危険箇所、テスト不足、例外・失敗時、既存設計整合性、セキュリティ・性能の順で見落としを潰す。
4. 退行、未解消リスク、仕様逸脱、未検証を優先して見る。bugfix scope 未達、退行、新規の仕様逸脱を severity 判定の中心に置く。
5. `severity` は未解消の最高重大度を返す。`critical` `medium` `low` の 3 段階を使い、好みや任意改善は指摘理由に含めない。
6. `score` は「欠陥の重さ」を表す離散バンドで返す。判定優先順位は `critical > medium > verification不足 > low件数 > external noise` とし、複数条件がある場合は最も低いバンドを採用する。
7. `score` は以下の rubric に従う。
   - `1.00`: 未解消の品質欠陥なし。required verification も満たしている。
   - `0.90`: 未解消が `low` 1-2 件のみ、または `external_validation_noise` / `known_pre_existing_issue` のみ。
   - `0.85`: 未解消が `low` 3-4 件のみ。
   - `0.75`: 未解消の `medium` が 1 件以上ある、required verification 不足がある、または `low` が 5 件以上ある。
   - `0.50`: 未解消の `critical` が 1 件以上ある。
8. `references/templates.md` の `## review feedback` を唯一の schema 正本として、`score` `severity` `location` `violated_contract` `required_delta` `recheck` `docs_sync_needed` をその順で返す。
9. `score < 0.85` の場合は `required_delta` を欠落させず、review loop を継続できる形で返す。`required_delta` と `recheck` の本文では、未解消 scope、external validation noise、residual risk を区別して書く。
10. `docs_sync_needed` は score に影響させず、docs handoff 判断専用で返す。
11. 次工程の起動は行わず、結果だけを `fix-direction` へ返す。

## 出力形式
- `references/templates.md` の `## review feedback` を唯一の schema 正本として扱う
- field は `score` `severity` `location` `violated_contract` `required_delta` `recheck` `docs_sync_needed` の 7 個で固定する
- `score` は `0.0 - 1.0` の範囲で返す
- `required_delta` には `scope_failures` `external_validation_noise` `known_pre_existing_issue` を区別して書く
- `recheck` には rerun コマンドと residual risk を区別して書く
- 正本は `changes/<id>/context_board/fix-review.feedback.json` とし、packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行して `fix-review.feedback.validation.json` を出力する
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する

## `fix-direction` が判断する条件
- `score >= 0.85` の review だけを次工程へ渡してよい
- `score < 0.85` の場合は `required_delta` を返して review loop を継続する
- ただし external validation noise だけが残る場合は、`score = 0.90` を上限に `fix-direction` が residual risk として扱う余地を残す
- `low` のみでも 5 件以上ある場合は `score = 0.75` とし、review loop を継続する

## 許可される動作
- 実装方針の好みより退行と未解消リスクを優先する
- green 判定は、必須レビュー観点を一通り確認した後に行う
- `score >= 0.85` とするのは、`critical` / `medium` が未解消でない場合に限る
- `external_validation_noise` と `known_pre_existing_issue` は通常欠陥より軽く扱うが、score 上限は `0.90` とする
- `docs_sync_needed` は score 減点理由ではなく handoff 判断材料として扱う
- read-only として振る舞う
- 返却内容は review 結果に限り、次工程起動は `fix-direction` に委ねる
