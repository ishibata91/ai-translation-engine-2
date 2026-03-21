---
name: investigation-explorer
description: AI Translation Engine 2 専用。investigation-distill で特定されたポインタをもとに実際にコードを走査し、詳細な調査を行う。
---

# Investigation Explorer

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `investigation-explorer` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は、渡されたファイルやシンボルのポインタを辿って、コードの詳細な構造や仕様、ロジックの繋がりを解き明かすための skill。

## 制約
- 明示的に指示されない限り、コードの書き換え（修正）は行わない。
- 調査対象外の広範な探索に逸脱しないよう、渡されたポインタを中心に探索する。

## やること
1. `investigation-direction` （および `investigation-distill` の結果）から渡された調査目的とポインタ（ファイル群・シンボル群）を確認する。
2. ポインタで指定されたファイルを詳細に読み解く（`view_file` や `server-filesystem` MCP のファイル読み取りなどを使用）。
3. 必要に応じて関数やメソッドの呼び出し元、呼び出し先を `go-llm-lens` や grep 検索等で辿り、文脈を構築する。
4. 調査の目的に対する回答としての事実、仕様、ロジックの流れを詳細に整理したレポートを作成し、`investigation-direction` 側に返す。

## 原則
- 憶測と事実を明確に区別し、ファイル名や行番号などの証拠を伴って報告する。
- わからなかったこと（不足しているコンテキスト）があれば、適当に補完せず「不明点」として正確に報告する。
