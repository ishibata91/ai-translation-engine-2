---
name: impl-direction
description: AI Translation Engine 2 専用。実装依頼、UI 反映、frontend / backend task 着手の入口整理だけを行うユーザー向け direction skill。plan artifacts の充足確認、実装フロー開始、review 後の handoff 管理に使い、蒸留、task 分割、実装、レビューは直接行わない。自由文の意図が設計や bugfix なら停止して適切な direction skill へ handoff するときにも使う。
---

# Impl Direction

この skill は impl 系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
自分では蒸留、task 分割、実装、レビューを行わず、artifact 充足確認と次に起動する skill / agent の決定だけを行う。

## 使う場面
- plan artifact が揃っているか確認したい
- 実装へ進めるか plan 系へ戻すか決めたい
- `impl-distill` `impl-workplan` `impl-review` の順番を管理したい
- review 結果に応じて worker へ差し戻すか `plan-sync` へ handoff するか決めたい

## 入口制約
- ユーザーから直接受けてよいのは実装、UI 反映、frontend / backend task 着手、plan artifact 充足確認だけとする。
- `impl-distill` `impl-workplan` `impl-frontend-work` `impl-backend-work` `impl-review` のような non-direction skill の直指定は受け付けず、`impl-direction` へ戻す handoff を返して停止する。
- 自由文が設計、仕様補完、docs 同期、bugfix、再現、原因調査を要求している場合は conflict として扱い、処理を進めず適切な direction skill を返して停止する。

## Conflict Policy
- `impl-direction` の正当入力 lane は `実装 / UI 反映 / frontend task / backend task / artifact readiness` とする。
- 自由文の意図に `設計 / 仕様補完 / docs 同期 / artifact 不足整理` が含まれる場合は `plan-direction` へ handoff する。
- 自由文の意図に `不具合 / 再現 / 原因切り分け / 修正方針整理` が含まれる場合は `fix-direction` へ handoff する。
- conflict を検出したら `impl-distill` 以降へ進まず、`references/templates.md` の conflict template で返答して終了する。

## agent / skill 対応
- 実装 packet の蒸留は `ctx_loader` に `impl-distill` を使わせる
- task 分割と worker 起動は `implementer` に `impl-workplan` を使わせる
- 実装は `sub_implementer` に `impl-frontend-work` または `impl-backend-work` を使わせる
- レビューは `review_cycler` に `impl-review` を使わせる

## 手順
1. 依頼が impl lane に属するかを先に判定する。non-direction skill の直指定や plan / fix lane の要求が含まれるなら conflict として停止する。
2. 実装依頼と `changes/<id>/` の artifact を読む。
3. `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無、鮮度、矛盾を確認する。
4. 不足、古さ、矛盾がある場合は `plan-distill` を起点に plan 系へ handoff する。
5. 実装に進める場合だけ `impl-distill` を起動する。
6. `impl-distill` を起動した後は implementation packet を待つ。返却前に自分で追加走査、追加読解、worker 選定の先行実施を始めない。
7. `impl-distill` が unknowns なしの implementation packet を返したら `impl-workplan` を起動する。
8. 実装完了後、統合差分を対象に `impl-review` を起動する。
9. `impl-review` が required delta を返すか `score < 0.85` の場合は、該当 worker へ差し戻す。
10. `score >= 0.85` を満たし、`docs_sync_needed` が true の場合だけ `plan-sync` へ handoff する。

## 標準チェーン
- 標準: `impl-direction` -> `impl-distill` -> `impl-workplan` -> `impl-frontend-work / impl-backend-work` -> `impl-review`
- plan へ戻す例外: `impl-direction` -> `plan-distill` -> `plan-direction` -> `plan-*` -> `plan-review` -> `impl-direction`

## 参照資料
- 起動テンプレートと handoff 形式は `references/templates.md` を使う。
- frontend / backend / mixed の見分け方は `references/quality-checklist.md` を読む。
- routing 例は `references/examples.md` を読む。

## 原則
- 指揮役は `orchestration-only` として振る舞い、蒸留、task 分割、実装、レビューを直接行わない
- agent 選択は `.codex/agents` を正本にする
- plan artifact が曖昧なまま `impl-workplan` へ進めない
- review feedback を要約せず、必要な worker へそのまま返す
- `score >= 0.85` を満たさない review では次工程へ進めない
- distill 起動後は packet を待ち、自分で追加走査、追加読解、worker 起動を行わない
- implementation packet が不足しているなら自分で読むのではなく `impl-distill` を再実行する
- conflict を検出したら自動補正や自動続行を行わず、停止して handoff だけを返す
