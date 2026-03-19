---
name: aite2-design-orchestrator
description: AI Translation Engine 2 専用。最初に使う design skill を決める。「どこから設計を始めるべきか分からない」と言われたときに起動する。
---

# AITE2 Design Orchestrator

この skill は design 系 skill の入口を整理し、最初に使う skill と次の 1 手を決めるための司令塔 skill。

## 使う場面
- どの design skill から始めるべきか迷う
- UI、シナリオ、ロジックのどこが未確定かを整理したい
- 既存 UI の見た目修正か、新規設計かを切り分けたい
- change 文書から docs 反映までの流れを整理したい
- 設計レビューと docs 同期まで含めた標準チェーンを決めたい

## 必読 spec
- `docs/governance/spec-structure/spec.md`
- `docs/governance/architecture/spec.md`

## 手順
1. `docs/governance/spec-structure/spec.md` で文書の置き場と責務区分を確認する。
2. `docs/governance/architecture/spec.md` で責務境界を確認する。
3. 要求が既存見た目修正か、新規設計か、docs 同期かを分類する。
4. UI、シナリオ、ロジックのうち未確定な層を特定する。
5. 最初に使う skill を 1 つ決める。
6. 次に使う skill の順番を決める。
7. 各ステップで確定事項、未確定事項、次の 1 手を明確にする。

## 標準チェーン
- UI を含む新規設計: `aite2-ui-design` -> `aite2-scenario-design` -> `aite2-logic-design` -> `aite2-design-review-guard` -> `aite2-sync-docs`
- UI を含まない backend / 責務設計: `aite2-logic-design` -> `aite2-design-review-guard` -> `aite2-sync-docs`
- 既存見た目修正: `aite2-ui-polish`
- docs だけを更新したい: `aite2-sync-docs`

## 設計と実装の分離
- 設計 skill から implementation skill へは自動遷移しない
- 実装へ進むかどうかはユーザーの明示指示で決める
- 設計が固まった後は、まず `aite2-design-review-guard` で設計レビューを行う
- docs 正本へ上げる必要がある場合だけ `aite2-sync-docs` を使う

## 参照資料
- 振り分け例は `references/routing-examples.md` を読む。
- 順序判断は `references/sequence-checklist.md` を使う。
- `scripts/init-change-design-docs.ps1` の使いどころは `references/doc-init-notes.md` を読む。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 自分で深い設計を書き切らず、適切な skill へ振り分ける
- UI、シナリオ、ロジックを一度に確定しようとしない
- 各ステップで確定事項、未確定事項、次の 1 手を明確にする
- `logic.md` を書いただけで完了扱いにしない
- backend / 責務変更がある設計は `aite2-design-review-guard` を標準で後続に置く
- 必要なときだけ `scripts/init-change-design-docs.ps1` を使う
