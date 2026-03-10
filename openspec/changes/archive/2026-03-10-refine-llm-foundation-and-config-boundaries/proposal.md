## Why

現在の `telemetry` と `progress` は `runtime` 配下に置かれている一方で、実際には `controller`、`workflow`、`slice`、`gateway` から横断的に利用されている。特に LLM 周辺では `pkg/gateway/llm/**` が `pkg/runtime/telemetry` に直接依存しており、`architecture.md` の依存方向と実装が衝突しているため、横断基盤としての `foundation` を新設して境界を整理する必要がある。

## What Changes

- `architecture.md` に `foundation` 区分を追加し、`telemetry` と `progress` を runtime 固有基盤ではなく横断基盤として扱う方針を明文化する。
- `backend-quality-gates` の `depguard` ルールを更新し、`foundation` への許可依存と `runtime` からの切り離しを両立させる。
- LLM 周辺に限定して `telemetry` と `progress` の import を `foundation` 配下へ移し、`pkg/gateway/llm/**`、関連する `runtime` / `workflow` / `controller` の境界違反を解消する。
- gateway が返す中立 DTO と、上位層が持つ進行解釈の責務分離を維持したまま、LLM 向けの通知・観測基盤だけを再配置する。
- **BREAKING** `pkg/runtime/telemetry` と `pkg/runtime/progress` を直接 import している箇所は、新しい `foundation` 配下へ移行する。

## Capabilities

### New Capabilities
- `foundation-boundary`: telemetry / progress のような横断基盤を `foundation` 区分として扱い、全責務区分から利用できる境界を定義する

### Modified Capabilities
- `backend-quality-gates`: `depguard` が foundation 区分を考慮し、runtime 固有依存と横断基盤依存を区別して検出するよう更新する
- `progress`: progress notifier を runtime 固有ではなく横断基盤として利用できるよう要件を更新する
- `telemetry`: telemetry を runtime 固有ではなく横断基盤として利用できるよう要件を更新する

## Impact

- 影響範囲は `openspec/specs/architecture.md`、`.golangci.yml`、`pkg/runtime/telemetry`、`pkg/runtime/progress`、および LLM 周辺でこれらを利用する `pkg/gateway/llm/**`、`pkg/runtime/**`、`pkg/workflow/**`、`pkg/controller/**`。
- スコープは LLM 周辺の import と provider wiring に限定し、dictionary や parser など他ユースケースへの全面展開はこの change では扱わない。
- Wails ログ連携や progress event 名は維持しつつ、package 依存だけを整理する移行が必要になる。
