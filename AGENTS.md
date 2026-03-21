# AI Assistant ルール定義

このファイルはこのプロジェクト向けの常設ルールです。毎回の指示を省くため、以下に従ってください。

## 出力言語
- 返答、資料、プランはすべて日本語で出力する。

## 役割確認
- すべての agent は、まず「このタスクに自分が適切か」を確認すること。
- 適切でない場合は、作業を進めず、オーケストレータ向けの handoff 提案を返すこと。

## ユーザー向け入口
- ユーザー向け入口として扱ってよい skill は `plan-direction` `impl-direction` `fix-direction` `investigation-direction` の 4 本だけとする。
- `plan-ui` `impl-frontend-work` `fix-work` など、non-direction skill の直指定は受け付けない。
- non-direction skill が指定された場合は実行せず、対応する direction skill への handoff を返して停止する。
- 明示された direction skill と自由文の意図が衝突した場合は downstream work を始めず、競合理由と正しい direction skill を返して停止する。
- 入口の使い分けは次の通りとする。
  - 設計、仕様補完、docs 同期は `plan-direction`
  - 実装、UI 反映、task 着手は `impl-direction`
  - 不具合、再現、原因切り分けは `fix-direction`
  - コード/仕様調査は `investigation-direction`

## 進め方
- `go-llm-lens` が入っている。`pkg/` 以下を走査したいときは使うこと。
- `ts-lsp` が入っている。`frontend/src/` 以下を走査したいときは使うこと。
- `ts-lsp` を使うときは、`projectRoot` に `F:/ai translation engine 2/frontend` のような実体の絶対パスを渡すこと。
- `ts-lsp` の `file` 指定も、必要に応じて絶対パスを使うこと。
- `server-filesystem` MCP が入っている。検索、読み書き、参照は必ずこれを利用すること。
- 書き込みも原則として `server-filesystem` を利用すること。
