---
name: aite2-design-orchestrator
description: AI Translation Engine 2 専用。設計要求を見て、ui-design、scenario-design、logic-design、design-fix、sync-docs のどれから始めるかを決め、1 ステップずつ進めるための統括 skill。設計の入口整理と順序制御が必要なときに使う。
---

# AITE2 Design Orchestrator

この skill は design 系 skill の司令塔用。
目的は、設計要求を見て最初に使うべき skill と進行順を決め、対話内で 1 ステップずつ進めること。

## 使う場面
- どの design skill から始めるべきか迷う
- UI、シナリオ、ロジックのどこが未確定かを整理したい
- 既存 UI の見た目修正か、新規設計かを切り分けたい
- change 文書から docs 反映までの流れを整理したい

## 役割
- 変更要求を分類する
- 最初に使う skill を決める
- 次に使う skill の順番を決める
- 対話内でタスクを 1 ステップずつ切る
- どこまで確定し、何が未確定かを整理する

## 基本順序
- UI を含む新規変更:
  - `aite2-ui-design`
  - `aite2-scenario-design`
  - `aite2-logic-design`
- UI を含まない変更:
  - `aite2-scenario-design`
  - `aite2-logic-design`
- 既存見た目修正:
  - `aite2-design-fix`
- 正本 docs への反映:
  - `aite2-sync-docs`

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 自分で深い設計を書き切らず、適切な skill へ振り分ける
- UI、シナリオ、ロジックを一度に確定しようとしない
- 各ステップで確定事項、未確定事項、次の 1 手を明確にする

## scripts
`scripts/init-change-design-docs.ps1` は `changes/<id>/` に `ui.md` `scenarios.md` `logic.md` のテンプレートを配置する。
引数で配置対象を選べる。
