---
name: aite2-bug-fix
description: AI Translation Engine 2 専用。バグを修正し、仕様との乖離を検知する。「〇〇を修正して」「バグ:〇〇が〇〇しない」と言われたときに起動する。
---

# AITE2 Bug Fix

この skill は bugfix 系の指揮者として動き、`aite2-explorer` による context 圧縮、`aite2-debugger` による原因絞り込み、ユーザー再現、`aite2-fixer` への修正指示、`aite2-implementation-review-guard` による review までを管理する orchestration skill。

## 使う場面
- 「〇〇を修正して」と既存機能の挙動修正を依頼された
- 「バグ: 〇〇が〇〇しない」と不具合報告を受けた
- 実装と `docs/` のどちらが正か切り分ける必要がある
- 回帰不具合、再現手順付き不具合、UI 上の観測差分を扱う

## subagent 起動規約
- context 収集は `aite2-explorer` を使う
- ログ構造化と事実抽出は `aite2-log-parser` を使う
- 原因予測、専用 logger 配置、原因の再絞り込みは `aite2-debugger` を使う
- 実際の修正は `aite2-fixer` を使う
- review は `aite2-implementation-review-guard` を使う
- handoff は `changes/<id>/context_board/` を通して行う
- debugger の一時ログは context board とは別に管理する

## 手順
1. change が無い場合は `scripts/init-change-bugfix-docs.ps1` で `changes/<id>/context_board/` を作る。
2. `aite2-explorer` に関連 spec、関連コード、既知の再現条件を集めさせ、board を初期化する。
3. 期待挙動と実挙動を board に分けて記録する。
4. `aite2-debugger` に原因仮説を立てさせ、専用 logger と専用出力を配置させる。
5. ユーザーが操作してバグを再現する。
6. `aite2-log-parser` に、取得したログや観測結果を構造化 JSON と board へ整理させる。
7. `aite2-debugger` に原因をさらに絞り込ませる。
8. bugfix 指揮者が、原因候補、仕様差分、修正対象を踏まえて実装プランを作る。
9. `aite2-fixer` に、指示された修正だけを実施させる。
10. 対象レイヤーに応じた品質ゲートを通す。
11. `aite2-implementation-review-guard` で実装レビューを行う。
12. finding があれば board に追記し、必要な skill へ差し戻す。

## 参照資料
- 具体例は `references/examples.md` を読む。
- 記録テンプレートは `references/templates.md` を使う。
- 仕様乖離の見分け方は `references/spec-gap-checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 再現、仕様確認、原因切り分け、修正、回帰確認を一度に混ぜない
- 各ステップで観測事実、判断、次の 1 手を明確にする
- 実装だけを見て正解を決めない
- 先に再現と仕様確認を行い、修正はその後に行う
- 修正より先に失敗条件を明確にする
- bugfix の handoff は context board に残し、口頭説明に依存しない
- UI を含む不具合でも、観測事実は logger と board に分けて残す
- 仕様が古い、誤っている、または不足している場合は、docs 修正とコード修正を同じ変更セットで扱う
- 仕様文書同士が矛盾する場合は、ユーザー確認なしに片方を採用しない
- 関連しないリファクタを混ぜない
- 必要な品質ゲートは省略しない
- 修正後の自己確認で止めず、`aite2-implementation-review-guard` を後続に置く
