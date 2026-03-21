---
name: impl-direction
description: AI Translation Engine 2 専用。実装依頼、UI 反映、frontend / backend task 着手の入口整理だけを行うユーザー向け direction skill。plan artifacts の充足確認、frontend/backend への task 分割と owned paths 確定、実装フロー開始、review 後の handoff 管理に使い、蒸留・実装・レビューは直接行わない。自由文の意図が設計や bugfix なら停止して適切な direction skill へ handoff する。
---

# Impl Direction

この skill は impl 系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
自分では蒸留・実装・レビューを行わず、以下だけを行う。

- artifact 充足確認
- frontend / backend / mixed の判定と task 分割
- owned paths / forbidden paths の確定
- shared contract の確定（mixed 実装時）
- 次に起動する skill / agent の選定

## 使う場面
- plan artifact が揃っているか確認したい
- 実装へ進めるか plan 系へ戻すか決めたい
- `impl-distill` → `implementer` → `impl-review` の順番を管理したい
- review 結果に応じて `implementer` へ差し戻すか `plan-sync` へ handoff するか決めたい

## 入口制約
- ユーザーから直接受けてよいのは実装、UI 反映、frontend / backend task 着手、plan artifact 充足確認だけとする。
- plan artifactは､ui.mdがない時はフロントエンド実装なしとして進める｡
- `impl-distill` `impl-frontend-work` `impl-backend-work` `impl-review` のような non-direction skill の直指定は受け付けず、`impl-direction` へ戻す handoff を返して停止する。
- 自由文が設計、仕様補完、docs 同期、bugfix、再現、原因調査を要求している場合は conflict として扱い、処理を進めず適切な direction skill を返して停止する。

## Conflict Policy
- `impl-direction` の正当入力 lane は `実装 / UI 反映 / frontend task / backend task / artifact readiness` とする。
- 自由文の意図に `設計 / 仕様補完 / docs 同期 / artifact 不足整理` が含まれる場合は `plan-direction` へ handoff する。
- 自由文の意図に `不具合 / 再現 / 原因切り分け / 修正方針整理` が含まれる場合は `fix-direction` へ handoff する。
- conflict を検出したら `impl-distill` 以降へ進まず、`references/templates.md` の conflict template で返答して終了する。

## agent / skill 対応
- 実装 packet の蒸留は `ctx_loader` に `impl-distill` を使わせる
- 実装は `implementer` agent に直接担当させる（`impl-frontend-work` / `impl-backend-work` は使わない）
- レビューは `review_cycler` に `impl-review` を使わせる

## 手順
1. 依頼が impl lane に属するかを先に判定する。non-direction skill の直指定や plan / fix lane の要求が含まれるなら conflict として停止する。
2. 実装依頼と `changes/<id>/` の artifact を読む。
3. `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無、鮮度、矛盾を確認する。
4. 不足、古さ、矛盾がある場合は `plan-distill` を起点に plan 系へ handoff する。
5. 実装に進める場合だけ `impl-distill` を起動する（`ctx_loader` agent を使う）。
6. `impl-distill` を起動した後は implementation packet を待つ。返却前に自分で追加走査・読解・agent 選定を始めない。
7. `impl-distill` の返却に unknowns、artifact 不足、曖昧な shared contract が含まれる場合は **task 分割を始めず**、不足解消に必要な観点を添えて `impl-distill` を再実行する。
8. `impl-distill` が unknowns なしの implementation packet を返したら **task 分割** を行う。
   - `references/quality-checklist.md` で frontend / backend / mixed を判定する。
   - worker ごとに **owned paths** と **forbidden paths** を決定する。
   - mixed 実装の場合は **shared contract**（型定義・API 契約）を `implementer` 起動前に 1 回だけ確定する。
   - 1 つの `implementer` 起動に frontend と backend の品質ゲートを混在させない。
   - 分割結果は `references/templates.md` の「task 分割」テンプレートで記録する。
9. 分割した task ごとに `implementer` agent を起動する。`implementer` は work order を受け取り、owned paths 内でコードを直接実装する。
10. 実装完了後、統合差分を対象に `impl-review` を起動する（`review_cycler` agent を使う）。
11. `impl-review` が required delta を返すか `score < 0.85` の場合は、該当 `implementer` へ差し戻す。
12. `score >= 0.85` を満たし、`docs_sync_needed` が true の場合だけ `plan-sync` へ handoff する。
13. `score >= 0.85` かつ `docs_sync_needed` が false の場合は impl lane 完了として終了する。

## 標準チェーン
- 標準: `impl-direction` -> `impl-distill` (ctx_loader) -> **[task 分割]** -> `implementer` -> `impl-review` (review_cycler)
- plan へ戻す例外: `impl-direction` -> `plan-distill` -> `plan-direction` -> `plan-*` -> `plan-review` -> `impl-direction`

## 終了条件
- review の `score >= 0.85` を満たし、impl lane で必要な修正 loop が完了している
- docs 昇格が必要な場合は `plan-sync` handoff を明示して終える
- docs 昇格が不要な場合は impl 完了として終える
- conflict の場合は downstream work を始めず、正しい direction skill を明示した handoff を返して終える

## 参照資料
- 起動テンプレートと handoff 形式は `references/templates.md` を使う。
- frontend / backend / mixed の見分け方は `references/quality-checklist.md` を読む。
- routing 例は `references/examples.md` を読む。

## 下流スキル起動時のスキル名明示
- `impl-distill` や `impl-review` をサブエージェントとして起動するときは、必ず `references/templates.md` の「下流スキル起動」テンプレートを使い、`invoked_skill` と `invoked_by` を明示すること
- `implementer` は skill ではなく agent であるため、「下流スキル起動」テンプレートの対象外とする
- `invoked_skill` には起動する下流スキル名（例: `impl-distill`）、`invoked_by` には `impl-direction` を設定する
- サブエージェントは起動時にこの情報で「自分がどのスキルとして起動されたか」を確認できる

## 原則
- 指揮役は `orchestration-only` として振る舞い、蒸留・実装・レビューを直接行わない
- agent 選択は `.codex/agents` を正本にする
- plan artifact が曖昧なまま `implementer` へ進めない
- implementation packet に unknowns や未確定 contract が残る状態で `implementer` へ進めない
- shared contract は混在 task でも `implementer` 起動前に必ず確定する（`implementer` に設計判断を残さない）
- review feedback を要約せず、必要な `implementer` へそのまま返す
- `score >= 0.85` を満たさない review では次工程へ進めない
- distill 起動後は packet を待ち、自分で追加走査・読解・agent 起動を行わない
- implementation packet が不足しているなら自分で読むのではなく `impl-distill` を再実行する
- conflict を検出したら自動補正や自動続行を行わず、停止して handoff だけを返す
