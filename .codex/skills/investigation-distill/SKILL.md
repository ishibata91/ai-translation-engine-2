---
name: investigation-distill
description: AI Translation Engine 2 専用。調査対象に関連しそうなファイル・パッケージ・シンボルを特定してリスト化（ポインタの蒸留）する。
---

# Investigation Distill
> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `investigation-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は調査に必要な関連ポインタ（ファイルパス、パッケージ名、シンボル名など）を洗い出すための skill。
調査のスコープを絞り込み、`investigation-direction` が次判断できるようにする。

## 許可される運用範囲
- 返却内容は「関連しそうな箇所（ポインタ）」の特定に限り、深い読解やロジック解析の判断は `investigation-direction` に委ねる。
- 恒久修正や設計提案は含めない。

## やること
1. ユーザーまたは `investigation-direction` からの調査リクエストの内容を確認する。
3. `pkg/` 以下を調べるときは `go-llm-lens` の package / symbol 系ツールを使って、関連パッケージやシンボルを特定する。。
6. 調査すべきファイル、パッケージ、シンボルのリスト（ポインタ集）を作成して返す。

## 許可される動作
- ポインタのリスト化に集中し、詳細な仕組みの解明判断は `investigation-direction` に委ねる。
- 対象が多すぎる場合は、優先度の高いものに絞るか、パターンの抽出を行う。
- 規約外の旧 tool 名や汎用テキスト検索指示は残さず、どの探索対象にどの MCP を使うかを文面で判別できるようにする。
