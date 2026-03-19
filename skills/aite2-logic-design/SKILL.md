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
