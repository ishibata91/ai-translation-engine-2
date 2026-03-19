---
name: aite2-scenario-design
description: AI Translation Engine 2 専用。操作シナリオを設計する。「操作フローを整理したい」「受け入れ条件を固めたい」と言われたときに起動する。
---

# AITE2 Scenario Design

この skill は UI と内部実装の間にあるユーザー体験の流れを整理し、受け入れ条件のぶれを減らすための skill。

## 使う場面
- 新機能の操作フローを決めたい
- 正常系だけでなく異常系も整理したい
- E2E や受け入れ条件の起点を作りたい
- resume / cancel / retry を含む変更を扱いたい

## 必読 spec
- 対象機能に対応する `docs/workflow/.../spec.md`
- 対象機能に対応する `docs/slice/.../spec.md`
- 補助: `docs/governance/spec-structure/spec.md`

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
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 一度に複数シナリオを同時確定しようとしない
- 各ステップで確定した flow、残課題、次の 1 手を明確にする
- シナリオは UI レイアウトではなく行動の流れとして書く
- 主要異常系を省略しない
- システム都合ではなくユーザーから見た結果で書く
- 受け入れ条件まで落とし切る
- Mermaid のロバストネス図を必須にする
