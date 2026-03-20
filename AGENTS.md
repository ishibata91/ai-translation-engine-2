# AI Assistant ルール定義

このファイルはこのプロジェクト向けの常設ルールです。毎回の指示を省くため、以下に従ってください。

## 出力言語
- 返答、資料、プランはすべて日本語で出力する。

## Routing Rule
Every agent MUST check: "Am I the right agent for this task?"
If not, surface a handoff suggestion to the orchestrator before proceeding.

## User-Facing Skill Entry
- ユーザー向け入口として扱ってよい skill は `plan-direction` `impl-direction` `fix-direction` の 3 本だけとする。
- `plan-ui` `impl-frontend-work` `fix-work` など non-direction skill の直指定は受け付けず、実行せずに対応する direction skill への handoff を返して停止する。
- 明示された direction skill と自由文の意図が衝突した場合は downstream work を始めず、競合理由と正しい direction skill を返して停止する。
- 設計 / 仕様補完 / docs 同期は `plan-direction`、実装 / UI 反映 / task 着手は `impl-direction`、不具合 / 再現 / 原因切り分けは `fix-direction` を入口にする。

## 進め方
- `go-llm-lensが入っている｡pkg/以下で走査したいときは使うこと`
- `server-filesystem` MCPが入っている｡検索､書き込み､読み取りは"""必ず"""これを利用すること｡
- 書き込みも原則`server-filesystem`を利用すること｡
