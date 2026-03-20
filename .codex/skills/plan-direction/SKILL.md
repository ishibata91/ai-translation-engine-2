---
name: plan-direction
description: AI Translation Engine 2 専用。設計依頼の入口整理、routing、次に起動する agent と plan skill の決定を行う。差分仕様、UI 振る舞い、シナリオ、ロジック、設計レビュー、docs 同期の入口を決めたいときに使う。
---

# Plan Direction

この skill は plan 系作業の入口指揮を担当する。
自分では設計 artifact を作らず、依頼の整理、artifact 充足確認、次に使う skill と agent の決定だけを行う。

## 使う場面
- どの plan skill から始めるべきか迷う
- 差分仕様と実装計画の不足箇所を整理したい
- UI、シナリオ、ロジックのどこを先に確定すべきか決めたい
- 設計レビューや docs 同期まで含めた順番を決めたい

## 必読 spec
- `docs/governance/spec-structure/spec.md`
- `docs/governance/architecture/spec.md`

## agent / skill 対応
- 判断材料の蒸留は `ctx_loader` に `plan-distill` を使わせる
- UI / シナリオ / ロジックの作成は `spec_drafter` に `plan-ui` `plan-scenario` `plan-logic` を使わせる
- 設計レビューは `review_cycler` に `plan-review` を使わせる
- docs 同期は `spec_syncer` に `plan-sync` を使わせる

## 手順
1. `docs/governance/spec-structure/spec.md` で文書の置き場と責務区分を確認する。
2. `docs/governance/architecture/spec.md` で責務境界を確認する。
3. `plan-distill` を起動し、要求、既存 artifact、関連 spec、未確定論点を planning packet に蒸留させる。
4. 依頼が新規設計、設計補完、既存 UI 調整、docs 同期のどれかを分類する。
5. UI、シナリオ、ロジックのうち先に確定すべき層を 1 つ決める。
6. 必要な `plan-*` skill を順番に並べる。
7. 設計 artifact が揃ったら `plan-review` を挟む。
8. docs 正本へ昇格すべき差分がある場合だけ `plan-sync` を起動する。

## 標準チェーン
- UI を含む新規設計: `plan-distill` -> `plan-ui` -> `plan-scenario` -> `plan-logic` -> `plan-review` -> `plan-sync`
- UI を含まない責務設計: `plan-distill` -> `plan-logic` -> `plan-review` -> `plan-sync`
- docs だけを更新したい: `plan-distill` -> `plan-sync`
- 既存 UI の見た目修正: `plan-distill` -> `aite2-ui-polish-orchestrate`

## 終了条件
- 必要な plan artifact と review 結果が揃い、次に進む skill が一意に決まっている
- 実装へ進む場合は `impl-direction` へ渡す handoff を明示して終える

## 参照資料
- 振り分け例は `references/routing-examples.md` を読む。
- 順序判断は `references/sequence-checklist.md` を使う。
- `scripts/init-change-design-docs.ps1` の使いどころは `references/doc-init-notes.md` を読む。

## 原則
- 指揮役は orchestration 以外を行わない
- agent 選択は `.codex/agents` を正本にする
- plan artifact が不足したまま implementation へ進めない
- 各 handoff で確定事項、未確定事項、次の 1 手を明示する
