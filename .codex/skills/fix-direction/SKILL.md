---
name: fix-direction
description: AI Translation Engine 2 専用。障害報告の入口整理、再現情報の収集方針、次に起動する agent と fix skill の決定を行う。調査から修正レビューまでの bugfix flow を指揮したいときに使う。
---

# Fix Direction

この skill は fix 系作業の入口指揮を担当する。
再現条件の整理、distill、trace、analysis、fix 実装、review の順番を管理する。

## agent / skill 対応
- 初期事実の蒸留は `ctx_loader` に `fix-distill` を使わせる
- 原因仮説と観測計画は `fault_tracer` に `fix-trace` を使わせる
- ログ / 観測出力の整理は `ctx_loader` に `fix-analysis` を使わせる
- 修正実装は `sub_implementer` に `fix-work` を使わせる
- 修正レビューは `review_cycler` に `fix-review` を使わせる

## 制約
- 指揮役は orchestration 以外を行わない
- 再現前に恒久修正へ進めない
- handoff は `changes/<id>/context_board/` を通して残す

## やること
1. change が無ければ `scripts/init-change-bugfix-docs.ps1` で `changes/<id>/context_board/` を作る。
2. `fix-distill` を起動し、再現条件、関連仕様、関連コード、既知観測を bugfix packet に蒸留させる。
3. `fix-trace` を起動し、原因仮説と観測計画を作らせる。
4. 再現後に `fix-analysis` を起動し、ログと観測出力を事実へ圧縮させる。
5. fix scope が確定したら `fix-work` を起動する。
6. 実装後に `fix-review` を起動し、退行と未解消リスクを確認する。
7. docs 反映が必要な場合だけ `plan-sync` へ handoff する。

## 参照
- 詳細例は `references/examples.md` を使う。
- 記録テンプレートは `references/templates.md` を使う。
- 仕様乖離の見分け方は `references/spec-gap-checklist.md` を使う。
