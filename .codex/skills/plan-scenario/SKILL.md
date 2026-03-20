---
name: plan-scenario
description: AI Translation Engine 2 専用。操作シナリオ、主要分岐、受け入れ条件を設計する。ユーザーフローと acceptance を差分仕様として確定したいときに使う。
---

# Plan Scenario

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `plan-scenario` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill はユーザー体験の流れと受け入れ条件を `scenarios.md` に落とす skill。

## 使う場面
- 新機能の操作フローを決めたい
- 正常系だけでなく異常系も整理したい
- E2E や受け入れ条件の起点を作りたい
- resume / cancel / retry を含む変更を扱いたい

## 必読 spec
- 対象機能に対応する `docs/workflow/.../spec.md`
- 対象機能に対応する `docs/slice/.../spec.md`
- 補助: `docs/governance/spec-structure/spec.md`

## 入力契約
- `plan-direction` または `plan-distill` から渡された planning packet
- 対象 change
- 対象シナリオ
- `changes/<id>/scenarios.md` に落とすための決定事項

## 手順
1. 対象機能に対応する `docs/workflow/.../spec.md` と `docs/slice/.../spec.md` を特定して読む。
2. 必要なら `docs/governance/spec-structure/spec.md` で zone と責務区分を確認する。
3. ユーザー目的と Trigger を決める。
4. Preconditions を整理し、Main Flow を書く。
5. Alternate、Error、Empty、Resume / Retry / Cancel を洗い出す。
6. ロバストネス図で Actor、Boundary、Control、Entity を配置する。
7. Acceptance Criteria と E2E 観点へ落とす。
8. `changes/<id>/scenarios.md` の形へまとめる。

## 参照資料
- シナリオ文書の雛形は `references/templates.md` を使う。
- シナリオ記法の例は `references/examples.md` を読む。
- ロバストネス図の確認項目は `references/robustness-checklist.md` を使う。

## 原則
- シナリオは UI レイアウトではなく行動の流れとして書く
- 主要異常系を省略しない
- 受け入れ条件まで落とし切る
- 実装判断や worker 分割へ進まない
