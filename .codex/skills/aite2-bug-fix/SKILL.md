---
name: aite2-bug-fix
description: AI Translation Engine 2 専用。バグを修正し、仕様との乖離を検知する。「〇〇を修正して」「バグ:〇〇が〇〇しない」と言われたときに起動する。
---

# AITE2 Bug Fix

## 目的
- bugfix 全体を orchestration し、explorer、debugger、log-parser、fixer、review-guard の handoff を管理する。

## 制約
- bugfix 指揮者は orchestration 以外を行ってはいけない。
- 自分で探索、調査、実装、レビュー、検証をしてはいけない。
- 探索、調査、実装、レビュー、検証はすべて対応する subagent に委譲する。
- subagent 起動は必ず各 skill の `agents/openai.yaml` に準拠する。
- `agents/openai.yaml` が無い、または読めない場合は block として報告する。
- ローカル直実行は禁止とし、ユーザーが delegation を明示的に止めた場合のみ例外を相談する。
- handoff は `changes/<id>/context_board/` を通して行う。

## やること
1. change が無ければ `scripts/init-change-bugfix-docs.ps1` で `changes/<id>/context_board/` を作る。
2. `aite2-explorer` を起動し、context を圧縮させる。
3. explorer の返答を統合し、期待挙動と実挙動を board に記録する。
4. `aite2-debugger` を起動し、原因仮説と観測計画を作らせる。
5. ユーザー再現後、`aite2-log-parser` を起動し、ログ事実を整理させる。
6. `aite2-debugger` を再度起動し、原因を絞り込ませる。
7. fixer に渡す実装プランだけを作る。
8. `aite2-fixer` を起動し、修正を実施させる。
9. 必要な品質ゲートを通す。
10. `aite2-implementation-review-guard` を起動し、レビューさせる。
11. finding があれば board に追記し、必要な skill へ差し戻す。

## 参照
- 詳細例は `references/examples.md` を使う。
- 記録テンプレートは `references/templates.md` を使う。
- 仕様乖離の見分け方は `references/spec-gap-checklist.md` を使う。
