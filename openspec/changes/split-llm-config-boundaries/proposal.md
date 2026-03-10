## Why

現在の設定関連コードは `workflow/config` に集中しているが、そこには SQLite 実装、migration、TypedAccessor、persona slice に関わる prompt default が混在しており、責務境界が曖昧になっている。特に LLM 周辺では `gateway` が `workflow/config` を再エクスポートしており、gateway が workflow へ依存する構造違反を生んでいるため、設定境界を `configstore` / `configaccess` / `workflow/persona` に分離する必要がある。

## What Changes

- LLM 周辺に限定して、設定ストアの技術実装を `pkg/gateway/configstore` へ分離する方針を定義する。
- 実行時読取補助を `pkg/runtime/configaccess` へ分離し、TypedAccessor の責務を runtime 側へ寄せる。
- persona slice に関わる prompt default / 解釈を `pkg/workflow/persona` へ分離し、slice 中心で設定責務を読める形にする。
- gateway が `workflow/config` を再エクスポートする構造を廃止し、LLM 周辺の gateway / runtime / workflow が独立した contract を参照するよう整理する。
- **BREAKING** LLM 周辺で `pkg/workflow/config` や `pkg/gateway/config` を直接 import している箇所は、新しい package 構成へ移行する。

## Capabilities

### New Capabilities
- `config-boundary-split`: config store、runtime access、workflow 固有設定解釈を責務分離して配置する境界を扱う

### Modified Capabilities
- `config`: 設定永続化と設定解釈の責務を分離し、LLM 周辺が gateway / runtime / workflow の適切な境界から利用する要件へ更新する

## Impact

- 影響範囲は `pkg/workflow/config`、`pkg/gateway/config`、LLM 周辺の `pkg/gateway/llm/**`、`pkg/runtime/**`、`pkg/workflow/**`、`pkg/controller/**`。
- スコープは LLM 周辺で必要な設定読み書きに限定し、全機能の一斉 rename はこの change では扱わない。
- Wire provider、migration、default prompt 読み出し、secret access の import 更新が必要になる。
