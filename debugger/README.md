# Debugger Helpers

このディレクトリは bugfix 用の一時ロガー置き場です。

## 方針
- 本番ロガーや `pkg/` 配下の常設コードとは分離する
- import 追加と呼び出し追加だけで観測を仕込めるようにする
- 調査が終わったら `debugger/` import と呼び出しを一括削除して戻せる形にする

## Go
- package: `github.com/ishibata91/ai-translation-engine-2/debugger/go_debuglogger`
- file: `debugger/go_debuglogger/logger.go`
- output: JSONL
- env: `AITE2_DEBUGGER_EVENTS_PATH`

## TypeScript
- module: `frontend/src/debugger/debuggerLogger.ts`
- output: browser console と `window.__AITE2_DEBUGGER_LOGS__`
- cleanup: `createDebuggerLogger` import と call site を削除する
