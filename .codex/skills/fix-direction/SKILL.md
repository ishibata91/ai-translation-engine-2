---
name: fix-direction
description: AI Translation Engine 2 専用。障害報告、再現、原因切り分け、修正方針整理の入口だけを行うユーザー向け direction skill。調査から修正レビュー、必要時の docs handoff までの bugfix flow を指揮し、再現条件の確定、コードやログの読解、原因調査、fix 実装、review は下流 skill へ委譲する。自由文の意図が設計や通常実装なら停止して適切な direction skill へ handoff するときにも使う。
---

# Fix Direction

この skill は fix 系作業の入口指揮を担当する。
ユーザー向け入口として使ってよい direction skill の 1 つであり、`orchestration-only` で動作する。
下流 skill に再現条件の確定、原因調査、修正、review を委譲し、その順番と review loop の採否だけを管理する。

## agent / skill 対応
- 初期事実の蒸留は `ctx_loader` に `fix-distill` を使わせる
- 原因仮説と観測計画は `fault_tracer` に `fix-trace` を使わせる
- 観測ログの実装は `log_instrumenter` に `fix-logging` を使わせる（`fix-trace` がログ必要と判断した場合のみ）
- ログ / 観測出力の整理は `ctx_loader` に `fix-analysis` を使わせる
- 修正実装は `implementer` に `fix-work` を使わせる
- 修正レビューは `review_cycler` に `fix-review` を使わせる

## 入口制約
- ユーザーから直接受けてよいのは不具合、再現、原因切り分け、修正方針整理、仕様乖離確認だけとする。
- `fix-distill` `fix-trace` `fix-analysis` `fix-work` `fix-review` のような non-direction skill の直指定は受け付けず、`fix-direction` へ戻す handoff を返して停止する。
- 自由文が設計、仕様補完、docs 同期、通常実装、UI 反映、task 着手を要求している場合は conflict として扱い、処理を進めず適切な direction skill を返して停止する。

## Conflict Policy
- `fix-direction` の正当入力 lane は `不具合 / 再現 / 原因切り分け / 修正方針整理 / 仕様乖離確認` とする。
- 自由文の意図に `設計 / 仕様補完 / docs 同期 / artifact 不足整理` が含まれる場合は `plan-direction` へ handoff する。
- 自由文の意図に `実装 / UI 反映 / task 着手` が含まれる場合は `impl-direction` へ handoff する。
- conflict を検出したら downstream skill を起動せず、`references/templates.md` の conflict template で返答して終了する。

## 制約
- 指揮役は orchestration 以外を行わない
- 再現前に恒久修正へ進めない
- handoff はパケットとして直接引き渡す
- review の起動責務と `score` 判定は `fix-direction` だけが持つ
- docs-only の仕様乖離が確定した場合は code fix を起動せず `plan-direction` へ handoff する

## やること
1. 依頼が fix lane に属するかを先に判定する。non-direction skill の直指定や plan / impl lane の要求が含まれるなら conflict として停止する。
2. `fix-distill` を起動し、再現条件、関連仕様、関連コード、既知観測を bugfix packet に蒸留させる。
3. `fix-distill` を起動した後は bugfix packet を待つ。返却前に自分で追加走査、追加読解、fix scope の先行確定を始めない。
4. `fix-trace` を起動し、原因仮説と観測計画を作らせる。
5. `fix-trace` がログ追加を必要と判断した場合は `fix-logging` を起動し、観測ログをコードへ仕込ませる。
6. 追加観測が必要な場合は､ユーザーに再現依頼を出して終了合図を待つ｡
7. 追加観測が必要な場合だけ、再現後に `fix-analysis` を起動し、ログと観測出力を事実へ圧縮させる。
8. 調査結果が docs-only の仕様乖離に収束した場合は code fix を起動せず、`plan-direction` への handoff を返して終了する。
9. `fix-trace` または `fix-analysis` の返却で fix scope が確定したら `fix-work` を起動する。
10. 実装後に `fix-review` を起動し、退行と未解消リスクの評価結果を受け取る。
11. `fix-review` の `score < 0.85` の場合は `required_delta` を fix loop の入力として `fix-work` へ戻し、修正 loop を継続する。
12. `score >= 0.85` を満たした場合だけ、docs 反映が必要なら `plan-sync` へ handoff する。
13. `score >= 0.85` かつ docs handoff が不要なら bugfix 完了として終了する。
14. 手順の最後に、必ず `fix-logging` を再度起動し、観測用に追加した一時ログを削除させるよう依頼すること（オーケストレーター自身で自らログを削除したりコードを編集しないこと）。

## 参照
- 詳細例は `references/examples.md` を使う。
- 記録テンプレートは `references/templates.md` を使う。
- 仕様乖離の見分け方は `references/spec-gap-checklist.md` を使う。

## 下流スキル起動時のスキル名明示
- 下流スキルのサブエージェントを立ち上げるときは、必ず `references/templates.md` の「下流スキル起動」テンプレートを使い、`invoked_skill` と `invoked_by` を明示すること
- `invoked_skill` には起動する下流スキル名（例: `fix-distill`）、`invoked_by` には `fix-direction` を設定する
- サブエージェントは起動時にこの情報で「自分がどのスキルとして起動されたか」を確認できる

## 原則
- 指揮役は `orchestration-only` として振る舞い、distill / trace / analysis / fix 実装 / review を直接行わない
- docs、コード、ログの読解と原因調査は下流 skill へ委譲し、自分では行わない
- distill の返却前に自分で追加走査、追加読解、修正方針確定を始めない
- 追加読解が必要なら `fix-distill` または `fix-analysis` を再実行する
- review の起動、`score` の採否判定、`required_delta` による loop 継続は `fix-direction` が一元管理する
- `fix-analysis` は必須工程ではなく、trace と既知事実だけで scope が確定する場合は省略できる
- conflict を検出したら自動補正や自動続行を行わず、停止して handoff だけを返す
- `score >= 0.85` を満たさない review では次工程へ進めない
