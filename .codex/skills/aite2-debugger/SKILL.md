---
name: aite2-debugger
description: AI Translation Engine 2 専用。bugfix 調査で原因仮説を立て、専用 logger と観測出力を配置し、再現後の事実から原因を絞り込む。bugfix flow の debugger 役が必要なときに起動する。
---

# AITE2 Debugger

この skill は bugfix flow の debugger 役として、原因仮説、調査計画、専用 logger 配置、再現後の原因絞り込みを担当する。

## 使う場面
- bugfix flow で最初の原因仮説を立てたい
- 再現前に debugger 専用 logger を配置したい
- 再現後のログと観測事実から原因候補を絞り込みたい

## 入力契約
- bugfix 指揮者から渡された context board
- 関連する docs
- 関連コード
- 再現条件
- 構造化された観測事実

## subagent 実行前提
- この skill は bugfix 指揮者から subagent として起動される前提で使う
- subagent 起動時は `agents/openai.yaml` の profile 設定を使う

## 手順
1. context board から既知の症状と再現条件を読む。
2. 原因仮説と優先調査箇所を整理する。
3. `scripts/init-debugger-logger.ps1` で `changes/<id>/debugger_logs/` 配下に debugger 専用 logger と専用出力を配置する。
4. 再現後は構造化された観測事実を読み、原因候補を狭める。
5. bugfix 指揮者へ fix plan の前提となる原因整理を返す。

## 原則
- 恒久修正は行わない
- repo 常設 logger を汚さない
- 一時ログは後で一括削除しやすい形にする
- 原因仮説と観測事実を分けて返す

## 参照資料
- handoff には `references/templates.md` を使う。
- logger 初期化は `scripts/init-debugger-logger.ps1` と `scripts/debugger-logger.ps1` を使う。
- Go の一時ロガーは `debugger/go_debuglogger/logger.go` を使う。
- TypeScript の一時ロガーは `frontend/src/debugger/debuggerLogger.ts` を使う。
- どちらも import と call site を一括削除して戻せる形で使う。
