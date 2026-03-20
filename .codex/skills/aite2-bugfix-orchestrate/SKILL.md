---
name: aite2-bugfix-orchestrate
description: AI Translation Engine 2 専用。バグ修正の入口を整理し、調査から修正までの流れを指揮する。「〇〇を修正して」「バグ:〇〇が〇〇しない」と言われたときに起動する。
---

# AITE2 Bugfix Orchestrate

## 目的
- bugfix 全体を orchestration し、`aite2-context-collect`、`aite2-cause-investigate`、`aite2-log-analyze`、`aite2-bugfix-implement`、`aite2-implementation-review` の handoff を管理する。

## 制約
- bugfix 指揮者は orchestration 以外を行ってはいけない。
- 自分で探索、調査、実装、レビュー、検証をしてはいけない。
- 探索、調査、実装、レビュー、検証はすべて対応する subagent に委譲する。
- handoff は `changes/<id>/context_board/` を通して行う。

## やること
1. change が無ければ `scripts/init-change-bugfix-docs.ps1` で `changes/<id>/context_board/` を作る。
2. `aite2-context-collect` を起動し、context を圧縮させる。
3. context packet を統合し、期待挙動と実挙動を board に記録する。
4. `aite2-cause-investigate` を起動し、原因仮説と観測計画を作らせる。
5. ユーザー再現後、`aite2-log-analyze` を起動し、ログ事実を整理させる。
6. `aite2-cause-investigate` を再度起動し、原因を絞り込ませる。
7. `aite2-bugfix-implement` に渡す実装プランだけを作る。
8. `aite2-bugfix-implement` を起動し、修正を実施させる。
9. 必要な品質ゲートを通す。
10. `aite2-implementation-review` を起動し、レビューさせる。
11. finding があれば board に追記し、必要な skill へ差し戻す。

## 参照
- 詳細例は `references/examples.md` を使う。
- 記録テンプレートは `references/templates.md` を使う。
- 仕様乖離の見分け方は `references/spec-gap-checklist.md` を使う。
