---
name: impl-direction
description: AI Translation Engine 2 専用。実装依頼、UI 反映、frontend / backend task 着手の入口整理だけを行うユーザー向け direction skill。plan artifacts の充足確認、`impl-distill` と `impl-workplan` の起動、section 単位の実装フロー開始、review 後の handoff 管理に使い、蒸留・計画・実装・レビューは直接行わない。自由文の意図が設計や bugfix なら停止して適切な direction skill へ handoff する。
---

# Impl Direction

この skill は impl 系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
自分では蒸留・実装計画・実装・レビューを行わず、以下だけを行う。

- artifact 充足確認
- `impl-distill` と `impl-workplan` の順次起動
- `tasks.md` と progress summary の正本管理
- section ごとの dispatch 順と担当 skill の決定
- 次に起動する skill / agent の選定

## 使う場面
- plan artifact が揃っているか確認したい
- 実装へ進めるか plan 系へ戻すか決めたい
- `impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review` の順番を管理したい
- 長時間 change の resume / reroute で進捗を取り違えずに再開したい
- review 結果に応じて affected section へ差し戻すか `plan-sync` へ handoff するか決めたい

## 入口制約
- ユーザーから直接受けてよいのは実装、UI 反映、frontend / backend task 着手、plan artifact 充足確認だけとする。
- `ui.md` が無いことだけを根拠に backend-only と判定してはならず、`references/quality-checklist.md` の routing matrix で frontend 影響有無を判定する。
- `impl-distill` `impl-workplan` `impl-frontend-work` `impl-backend-work` `impl-review` のような non-direction skill の直指定は受け付けず、`impl-direction` へ戻す handoff を返して停止する。
- 自由文が設計、仕様補完、docs 同期、bugfix、再現、原因調査を要求している場合は conflict として扱い、処理を進めず適切な direction skill を返して停止する。

## Conflict Policy
- `impl-direction` の正当入力 lane は `実装 / UI 反映 / frontend task / backend task / artifact readiness` とする。
- 自由文の意図に `設計 / 仕様補完 / docs 同期 / artifact 不足整理` が含まれる場合は `plan-direction` へ handoff する。
- 自由文の意図に `不具合 / 再現 / 原因切り分け / 修正方針整理` が含まれる場合は `fix-direction` へ handoff する。
- conflict を検出したら `impl-distill` 以降へ進まず、`references/templates.md` の conflict template で返答して終了する。

## agent / skill 対応
- 実装 packet の蒸留は `ctx_loader` に `impl-distill` を使わせる
- section planning は `workplan_builder` に `impl-workplan` を使わせる
- section 実装は `implementer` agent に `impl-frontend-work` または `impl-backend-work` を使わせる
- レビューは `review_cycler` に `impl-review` を使わせる

## 手順
1. 依頼が impl lane に属するかを先に判定する。non-direction skill の直指定や plan / fix lane の要求が含まれるなら conflict として停止する。
2. 実装依頼と `changes/<id>/` の artifact を読む。
3. `ui.md` `scenarios.md` `logic.md` の有無、鮮度、矛盾を確認する。`tasks.md` が存在する場合は impl lane の progress source of truth として読む。
4. 不足、古さ、矛盾がある場合は `plan-distill` を起点に plan 系へ handoff する。
5. 実装に進める場合だけ `impl-distill` を起動する（`ctx_loader` agent を使う）。
6. `impl-distill` を起動した後は implementation packet を待つ。返却前に自分で追加走査・読解・agent 選定を始めない。
7. `impl-distill` の返却に unknowns、artifact 不足、曖昧な shared contract 候補が含まれる場合は **section planning を始めず**、不足解消に必要な観点を添えて `impl-distill` を再実行する。
8. `impl-distill` が unknowns なしの implementation packet を返したら `impl-workplan` を起動する（`workplan_builder` agent を使う）。
9. `impl-workplan` を起動した後は section plan、condensed brief、`changes/<id>/tasks.md` を待つ。返却前に自分で section 分割や worker 割り当てを始めない。
10. `impl-workplan` の返却に unresolved section、未固定 contract、owner 未確定が残る場合は **実装を始めず**、不足解消に必要な観点を添えて `impl-workplan` を再実行する。
11. `impl-workplan` が有効な section plan を返したら、`tasks.md` と progress snapshot を照合して resume reconciliation を行い、`pending` または reroute 指定された section だけを dispatch 候補にする。
12. dispatch payload は `references/templates.md` の `Section Dispatch` を使い、`title` `goal` `depends_on` `shared_contract` `required_reading` `validation_commands` `acceptance` に加えて `progress_snapshot` と `condensed_brief` を含む full work order schema をそのまま渡す。
   - `owner: frontend` の section は `implementer` agent に `impl-frontend-work` を使わせる。
   - `owner: backend` の section は `implementer` agent に `impl-backend-work` を使わせる。
   - 1 section = 1 owner を守り、1 つの section に frontend と backend の品質ゲートを混在させない。
13. section が `completed` または `blocked` を返したら、`references/templates.md` の `Section Result` を基に `tasks.md` の status、実装チェック、検証チェック、noise 注記を更新する。section 契約そのものは書き換えない。
14. section 結果が `blocked` で、原因が未固定 contract や progress snapshot 矛盾なら worker 再投入を行わず `impl-workplan` 再実行へ戻す。`external_validation_noise` または `known_pre_existing_issue` だけなら reroute 対象から外す。
15. section 結果の記録後、完了済み subagent は close し、progress summary の `next_dispatch` を更新する。
16. 全 section の実装完了後、統合差分を対象に `impl-review` を起動する（`review_cycler` agent を使う）。
17. `impl-review` が required delta を返すか `score < 0.85` の場合は、`affected_sections` を使って該当 section だけを再 dispatch する。
   - reroute では `impl-workplan` が確定した元の full section 契約を崩さず、`required_delta` に加えて `progress_snapshot` と `carry_over_contracts` を添えた状態要約 packet を渡す。
18. `score >= 0.85` を満たし、`docs_sync_needed` が true の場合だけ `plan-sync` へ handoff する。
19. `score >= 0.85` かつ `docs_sync_needed` が false の場合は impl lane 完了として終了する。

## 標準チェーン
- 正本 chain: `impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review`
- 標準: `impl-direction` -> `impl-distill` (ctx_loader) -> `impl-workplan` (workplan_builder) -> `impl-frontend-work` / `impl-backend-work` (`implementer`) -> `impl-review` (review_cycler)
- plan へ戻す例外: `impl-direction` -> `plan-distill` -> `plan-direction` -> `plan-*` -> `plan-review` -> `impl-direction`

## 終了条件
- review の `score >= 0.85` を満たし、impl lane で必要な修正 loop が完了している
- docs 昇格が必要な場合は `plan-sync` handoff を明示して終える
- docs 昇格が不要な場合は impl 完了として終える
- conflict の場合は downstream work を始めず、正しい direction skill を明示した handoff を返して終える

## 参照資料
- 起動テンプレートと handoff 形式は `references/templates.md` を使う。
- artifact readiness と section routing の見分け方は `references/quality-checklist.md` を読む。
- routing 例は `references/examples.md` を読む。

## 下流スキル起動時のスキル名明示
- `impl-distill` `impl-workplan` `impl-review` をサブエージェントとして起動するときは、必ず `references/templates.md` の「下流スキル起動」テンプレートを使い、`invoked_skill` と `invoked_by` を明示すること
- `impl-frontend-work` / `impl-backend-work` は `implementer` agent 上で動く execution skill だが、skill 名は必ず work order に明示する
- `invoked_skill` には起動する下流スキル名（例: `impl-workplan`）、`invoked_by` には `impl-direction` を設定する
- サブエージェントは起動時にこの情報で「自分がどのスキルとして起動されたか」を確認できる

## 原則
- 指揮役は `orchestration-only` として振る舞い、蒸留・計画・実装・レビューを直接行わない
- agent 選択は `.codex/agents` を正本にする
- plan artifact が曖昧なまま `impl-workplan` へ進めない
- implementation packet に unknowns や未確定 contract が残る状態で `impl-workplan` へ進めない
- section plan に未確定 owner や shared contract が残る状態で worker へ進めない
- `impl-workplan` を経ない worker 直 dispatch や、`tasks.md` をユーザー入力前提で読む旧フローを復活させない
- `tasks.md` は section 契約の正本ではなく progress の正本として扱い、section 契約変更は `impl-workplan` 以外で行わない
- `Workplan Summary` `Section Dispatch` `Review Reroute` の section schema から `shared_contract` `required_reading` `validation_commands` `acceptance` `condensed_brief` を省略しない
- review feedback を要約せず、必要な affected section へそのまま返す
- `score >= 0.85` を満たさない review では次工程へ進めない
- distill / workplan 起動後は packet を待ち、自分で追加走査・読解・section 分割を行わない
- implementation packet が不足しているなら自分で読むのではなく `impl-distill` を再実行する
- section plan が不足しているなら自分で埋めず、`impl-workplan` を再実行する
- worker の `completed_scope` `remaining_gap` `noise_classification` を見ずに reroute 判断しない
- completed subagent を開いたまま次 section を始めない
- conflict を検出したら自動補正や自動続行を行わず、停止して handoff だけを返す
