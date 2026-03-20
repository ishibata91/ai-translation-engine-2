---
name: aite2-design-orchestrator
description: AI Translation Engine 2 専用。最初に使う design skill を決める。「どこから設計を始めるべきか分からない」と言われたときに起動する。
---

# AITE2 Design Orchestrator

この skill は design 系 skill の司令塔として動き、`aite2-explorer` による context 収集、最初に使う設計 skill の決定、設計 review、docs sync までの順番を管理する orchestration skill。

## 使う場面
- どの design skill から始めるべきか迷う
- UI、シナリオ、ロジックのどこが未確定かを整理したい
- 既存 UI の見た目修正か、新規設計かを切り分けたい
- change 文書から docs 反映までの流れを整理したい
- 設計レビューと docs 同期まで含めた標準チェーンを決めたい

## 必読 spec
- `docs/governance/spec-structure/spec.md`
- `docs/governance/architecture/spec.md`

## subagent 起動規約
- design 指揮者は、原則として下位 skill を subagent として起動する
- 下位 skill の起動時は、各 skill 配下の `agents/openai.yaml` に定義された subagent 設定を使う
- ローカル直実行は、ユーザーが delegation を明示的に止めた場合か、subagent 側設定が欠けていて継続不能な場合に限る
- context 収集は `aite2-explorer` を使う
- 実際の design 作業は `aite2-ui-design` `aite2-scenario-design` `aite2-logic-design` を使う
- review は `aite2-design-review-guard` を使う
- docs sync が必要な場合は `aite2-sync-docs` を使う

## 手順
1. `docs/governance/spec-structure/spec.md` で文書の置き場と責務区分を確認する。
2. `docs/governance/architecture/spec.md` で責務境界を確認する。
3. `aite2-explorer` に、指示に関連する `changes/` `docs/` 既存コードの context packet を集めさせる。
4. 要求が既存見た目修正か、新規設計か、docs 同期かを分類する。
5. UI、シナリオ、ロジックのうち未確定な層を特定する。
6. 最初に使う design skill を 1 つ決める。
7. 次に使う design skill の順番を決める。
8. review と docs sync が必要か判定する。
9. 各ステップで確定事項、未確定事項、次の 1 手を明確にする。

## 標準 orchestration フロー
1. 指揮者は orchestration 以外を行わず、context 収集、routing、終了判定だけを行う。
2. 最初の context 収集は必須とする。
3. 設計作業は `aite2-ui-design` `aite2-scenario-design` `aite2-logic-design` `aite2-design-review-guard` `aite2-sync-docs` に委譲する。
4. context が足りなくなった場合だけ、次の design skill 起動前に `aite2-explorer` を再度呼ぶ。
5. 設計から implementation へは自動遷移しない。

## 標準チェーン
- UI を含む新規設計: `aite2-explorer` -> `aite2-ui-design` -> `aite2-scenario-design` -> `aite2-logic-design` -> `aite2-design-review-guard` -> `aite2-sync-docs`
- UI を含まない backend / 責務設計: `aite2-explorer` -> `aite2-logic-design` -> `aite2-design-review-guard` -> `aite2-sync-docs`
- 既存見た目修正: `aite2-explorer` -> `aite2-ui-polish`
- docs だけを更新したい: `aite2-explorer` -> `aite2-sync-docs`

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
- 自分で探索、設計、レビュー、同期を実行せず、適切な skill へ振り分ける
- UI、シナリオ、ロジックを一度に確定しようとしない
- 各ステップで確定事項、未確定事項、次の 1 手を明確にする
- `logic.md` を書いただけで完了扱いにしない
- backend / 責務変更がある設計は `aite2-design-review-guard` を標準で後続に置く
- 必要なときだけ `scripts/init-change-design-docs.ps1` を使う
- context packet を持たないまま下位 design skill を起動しない
