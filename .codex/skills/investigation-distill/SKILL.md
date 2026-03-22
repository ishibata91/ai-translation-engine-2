---
name: investigation-distill
description: AI Translation Engine 2 専用。調査対象に関連しそうなファイル・パッケージ・シンボルを特定してリスト化（ポインタの蒸留）する。
---

# Investigation Distill
> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `investigation-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は調査に必要な関連ポインタ（ファイルパス、パッケージ名、シンボル名など）を洗い出すための skill。
調査のスコープを絞り込み、後続の `investigation-explorer` が効率的に走査できるようにする。

## 制約
- コードの深い読解やロジックの解析は行わず、あくまで「関連しそうな箇所（ポインタ）」の特定にとどめること。
- 恒久修正や設計提案はしない。
- `AGENTS.md` の MCP 利用規約に従い、探索と参照は `server-filesystem` を正本とすること。
- `pkg/` 以下のパッケージやシンボル探索は `go-llm-lens` を正本とし、`frontend/src/` 以下の探索は `ts-lsp` を正本とすること。
- `ts-lsp` を使う場合は、`AGENTS.md` に従って `projectRoot` や `file` に絶対パスを渡すこと。

## やること
1. ユーザーまたは `investigation-direction` からの調査リクエストの内容を確認する。
2. パスやファイル候補の探索は `server-filesystem` のディレクトリ一覧、ファイル検索、テキスト読取で行い、関連ファイルを特定する。
3. `pkg/` 以下を調べるときは `go-llm-lens` の package / symbol 系ツールを使って、関連パッケージやシンボルを特定する。
4. `frontend/src/` 以下を調べるときは `ts-lsp` を使って、関連ファイル、型、参照位置を特定する。
5. 対象ファイルの一部を `server-filesystem` で必要最小限だけ読み、確度を評価する。
6. 調査すべきファイル、パッケージ、シンボルのリスト（ポインタ集）を作成して返す。

## 原則
- ポインタのリスト化に集中し、詳細な仕組みの解明は `investigation-explorer` に任せる。
- 対象が多すぎる場合は、優先度の高いものに絞るか、パターンの抽出を行う。
- 規約外の旧 tool 名や汎用テキスト検索指示は残さず、どの探索対象にどの MCP を使うかを文面で判別できるようにする。
