---
name: impl-workplan
description: AI Translation Engine 2 専用。implementation packet を独立 work order へ分解し、shared contract を確定して `sub_implementer` worker を起動する。mixed 実装や ownership 分割が必要なときに使う。
---

# Impl Workplan

この skill は implementation packet を work order に分解し、`sub_implementer` worker を起動する skill。
implementation packet を独立 work order に分解し、shared contract を確定したうえで worker を起動する。

## 使う場面
- frontend / backend / mixed を work order に分けたい
- shared contract を worker 起動前に確定したい
- ownership と forbidden paths を明文化したい
- worker の返却形式を統一したい

## 入力契約
- `impl-distill` が返した implementation packet
- 共有 contract の候補
- integration 後に満たす acceptance

## 手順
1. implementation packet を読み、frontend / backend / mixed を判定する。
2. worker ごとに owned paths と forbidden paths を決める。
3. shared contract を 1 回だけ確定する。
4. `impl-frontend-work` または `impl-backend-work` を使う work order を作る。
5. `sub_implementer` worker を起動し、返却された structured diff を集約する。
6. integration に必要な検証結果をそろえて `impl-review` へ渡す。

## 出力契約
- `work_order`: worker ごとの owned paths, forbidden paths, required reading, shared contract, validation commands
- `integration_summary`: 返却された structured diff の統合結果
- `review_input`: `impl-review` へ渡す差分、検証結果、前回 findings

## 参照資料
- work order の雛形は `references/templates.md` を使う。
- frontend / backend / mixed の見分け方は `references/quality-checklist.md` を読む。
- 典型例は `references/examples.md` を読む。

## 原則
- shared contract は worker 起動前に必ず確定する
- worker に設計判断を残さない
- frontend と backend の品質ゲートを 1 work order に混ぜない
- mixed 実装でも review は統合差分に対して 1 回だけ行う
