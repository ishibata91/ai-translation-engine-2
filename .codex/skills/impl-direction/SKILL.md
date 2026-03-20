---
name: impl-direction
description: AI Translation Engine 2 専用。実装依頼の入口整理、routing、次に起動する agent と impl skill の決定を行う。plan artifacts の充足確認、実装フロー開始、review 後の handoff を管理したいときに使う。
---

# Impl Direction

この skill は impl 系作業の入口指揮を担当する。
自分では蒸留、task 分割、実装、レビューを行わず、artifact 充足確認と次に起動する skill / agent の決定だけを行う。

## 使う場面
- plan artifact が揃っているか確認したい
- 実装へ進めるか plan 系へ戻すか決めたい
- `impl-distill` `impl-workplan` `impl-review` の順番を管理したい
- review 結果に応じて worker へ差し戻すか `plan-sync` へ handoff するか決めたい

## agent / skill 対応
- 実装 packet の蒸留は `ctx_loader` に `impl-distill` を使わせる
- task 分割と worker 起動は `implementer` に `impl-workplan` を使わせる
- 実装は `sub_implementer` に `impl-frontend-work` または `impl-backend-work` を使わせる
- レビューは `review_cycler` に `impl-review` を使わせる

## 手順
1. 実装依頼と `changes/<id>/` の artifact を読む。
2. `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無、鮮度、矛盾を確認する。
3. 不足、古さ、矛盾がある場合は `plan-distill` を起点に plan 系へ handoff する。
4. 実装に進める場合だけ `impl-distill` を起動する。
5. `impl-distill` が unknowns なしの implementation packet を返したら `impl-workplan` を起動する。
6. 実装完了後、統合差分を対象に `impl-review` を起動する。
7. `impl-review` が required delta を返したら、該当 worker へ差し戻す。
8. `docs_sync_needed` が true の場合だけ `plan-sync` へ handoff する。

## 標準チェーン
- 標準: `impl-direction` -> `impl-distill` -> `impl-workplan` -> `impl-frontend-work / impl-backend-work` -> `impl-review`
- plan へ戻す例外: `impl-direction` -> `plan-distill` -> `plan-direction` -> `plan-*` -> `plan-review` -> `impl-direction`

## 参照資料
- 起動テンプレートと handoff 形式は `references/templates.md` を使う。
- frontend / backend / mixed の見分け方は `references/quality-checklist.md` を読む。
- routing 例は `references/examples.md` を読む。

## 原則
- 指揮役は orchestration 以外を行わない
- agent 選択は `.codex/agents` を正本にする
- plan artifact が曖昧なまま `impl-workplan` へ進めない
- review feedback を要約せず、必要な worker へそのまま返す
