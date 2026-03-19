---
name: aite2-logic-design
description: AI Translation Engine 2 専用。ロジック設計を整理する。「責務分解を決めたい」「workflow と slice の境界を整理したい」と言われたときに起動する。
---

# AITE2 Logic Design

この skill はユーザー向け仕様を内部責務へ分解し、依存方向と保存境界を明確にするための skill。

## 使う場面
- backend を含む変更を設計したい
- workflow / runtime / gateway / artifact の分担に迷っている
- DTO や契約、保存先、依存方向を整理したい
- UI 要求をどこで実現するか切り分けたい

## 必読 spec
- `docs/governance/architecture/spec.md`
- 補助: `docs/governance/spec-structure/spec.md`

## 手順
1. `docs/governance/architecture/spec.md` を読み、責務境界と依存方向を確認する。
2. 必要なら `docs/governance/spec-structure/spec.md` で zone と文書の置き場を確認する。
3. 対象主シナリオを 1 つに絞る。
4. Goal と入力、出力、中間成果物を整理する。
5. Controller、Workflow、Slice、Artifact、Runtime、Gateway の責務を割り当てる。
6. Main Path と主要分岐を整理する。
7. Persistence Boundary と Side Effects を明確にする。
8. `changes/<id>/logic.md` の形へまとめる。
9. `docs` 正本へ昇格すべき仕様断片を洗い出す。
10. `aite2-design-review-guard` へ渡す論点を明示する。

## `logic.md` に含めるもの
- Goal、入力、出力、中間成果物
- Controller / Workflow / Slice / Artifact / Runtime / Gateway の責務分担
- Main Path / Key Branches
- Persistence Boundary / Side Effects
- Risks / 保留事項

## `logic.md` に含めるが docs へ丸ごと同期しないもの
- どこを正本にするかの設計理由
- 責務分担の比較検討
- 実装前の暫定メモ
- 実装都合の分解方針

## docs 昇格候補として抽出するもの
- 対象集合
- 恒久的な振る舞いルール
- 正常系 / 異常系の成立条件
- 重複統合や優先順位の規則
- 永続化される事実
- 外部公開契約として残る前提

## 次の skill
- `logic.md` 完了後は原則 `aite2-design-review-guard` を次の 1 手として案内する
- docs 正本へ上げる必要がある場合は、レビュー後に `aite2-sync-docs` を使う
- implementation skill へは自動遷移しない

## 参照資料
- logic 文書の雛形は `references/templates.md` を使う。
- 責務分解の例は `references/examples.md` を読む。
- 境界確認は `references/boundary-checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 責務分解、依存方向、保存境界を一度に確定しようとせず、論点ごとに進める
- 各ステップで確定事項、保留事項、次の 1 手を明確にする
- 1 文書 1 主シナリオを原則にする
- `logic.md` は実装手順書ではなく責務境界の割り当てに留める
- `logic.md` は change 内の判断材料であり、docs 正本へ同期する前提で膨らませない
- controller は進行決定しない
- workflow は orchestration に集中する
- runtime は外部 I/O 実行に集中する
- slice は個別業務ロジックに集中する
- artifact は handoff と共有成果物の保存境界に限定する
- 具象依存の拡散を避ける
