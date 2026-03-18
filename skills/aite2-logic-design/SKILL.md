---
name: aite2-logic-design
description: AI Translation Engine 2 専用。workflow、runtime、gateway、slice、artifact、controller、frontend の責務境界に沿ってロジック設計を整理する。内部フロー、契約、永続化、依存方向を決めるときに使う。
---

# AITE2 Logic Design

この skill は内部設計用。
目的は、ユーザー向け仕様を内部責務へ分解し、依存方向と保存責務を明確にすること。

## 使う場面
- backend を含む変更を設計する
- workflow / runtime / gateway / artifact の分担に迷う
- DTO や契約、保存先、依存方向を整理したい
- UI 要求をどこで実現するか切り分けたい

## 入力
- `changes/<id>/summary.md`
- 必要なら `changes/<id>/ui.md`
- 必要なら `changes/<id>/scenarios.md`
- 関連する docs と既存実装

## 出力
- `changes/<id>/logic.md`
- 責務分解
- 依存方向
- 変更対象モジュール一覧
- 必要なら `references/logic-template.md` を元にした logic 文書

## 粒度
- `logic.md` は 1 つの主シナリオを成立させるための責務分解を単位にする
- 画面全体、機能全体、パッケージ全体を一度に扱わない
- ボタン単位や関数単位まで細かくしない
- 別シナリオは別節または別文書に分ける

## 最低限含めること
- Scenario
- Goal
- Responsibility Split
- Data Flow
- Main Path
- Key Branches
- Persistence Boundary
- Side Effects
- Risks

## 書くこと
- どの責務区分が何を受け持つか
- 何を受け取り、何を返し、何を保存するか
- 正常系の主要経路
- 主要な異常系、resume、retry、cancel の分岐

## 書かないこと
- メソッド分解の詳細
- クラス設計の細部
- if 分岐の全列挙
- DTO フィールド一覧の詳細
- 局所的ワークアラウンド
- docs 正本へ同期すべき長期仕様

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 責務分解、依存方向、保存境界を一度に確定しようとせず、論点ごとに進める
- 各ステップで確定事項、保留事項、次の 1 手を明確にする
- 1 文書 1 主シナリオを原則にする
- `logic.md` は実装手順書ではなく責務境界の割り当てに留める
- `logic.md` は change 内の判断材料であり、docs 正本へ同期する前提で膨らませない
- `logic.md` の雛形は `references/logic-template.md` を使う
- controller は進行決定しない
- workflow は orchestration に集中する
- runtime は外部 I/O 実行に集中する
- slice は個別業務ロジックに集中する
- artifact は handoff と共有成果物の保存境界に限定する
- 具象依存の拡散を避ける

## 推奨構成
```md
# Logic Design

## Scenario
対象シナリオ名

## Goal
このロジックが成立させる業務目的

## Responsibility Split
- Controller:
- Workflow:
- Slice:
- Artifact:
- Runtime:
- Gateway:

## Data Flow
- 入力
- 中間成果物
- 出力

## Main Path
主経路の責務受け渡し

## Key Branches
主要分岐だけ

## Persistence Boundary
どこで何を保存するか

## Side Effects
外部 I/O、ジョブ投入、通知など

## Risks
崩れやすい点、未確定点
```

## テンプレート
`changes/<id>/logic.md` は `references/logic-template.md` を元に作る。
必要に応じて項目を追加してよいが、Scenario、Responsibility Split、Data Flow は省略しない。
