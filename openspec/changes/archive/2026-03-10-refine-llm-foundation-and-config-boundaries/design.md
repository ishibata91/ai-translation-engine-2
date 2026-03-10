## Context

`pkg/runtime/telemetry` と `pkg/runtime/progress` は命名上は runtime 固有基盤に見えるが、実際には `controller`、`workflow`、`slice`、`gateway` から横断的に利用されている。とくに LLM 周辺では `pkg/gateway/llm/**` が `pkg/runtime/telemetry` を直接 import しており、`architecture.md` の `gateway -> runtime` 禁止と衝突している。

この change では、横断基盤を `foundation` として切り出し、LLM 周辺だけを対象に `telemetry` と `progress` を再配置する。進行率の意味づけや DTO 解釈は `workflow` に残し、foundation には通知 transport と観測補助だけを置く。

## Goals / Non-Goals

**Goals:**
- `architecture.md` に `foundation` 区分を追加し、`telemetry` と `progress` を横断基盤として定義する。
- `depguard` が `foundation` を許可依存として扱えるように更新する。
- LLM 周辺に限定して `telemetry` / `progress` import を foundation 配下へ移す。
- `gateway` が上位層解釈を持たないまま、観測基盤だけを安全に利用できる状態へ寄せる。

**Non-Goals:**
- repository 全体の `telemetry` / `progress` import を一度に移行しない。
- `workflow` の phase 設計や progress 算出ルール自体を再設計しない。
- LLM 周辺以外の gateway / slice / controller の境界修正をこの change だけで完了させない。

## Decisions

### 1. `foundation` は横断基盤専用区分とする
- Decision:
  `telemetry` と `progress` を `foundation` 区分へ移し、`controller`、`workflow`、`slice`、`runtime`、`gateway` から参照できる共通基盤として扱う。
- Rationale:
  現実の利用形態と依存方向ルールを一致させるには、runtime 固有基盤ではなく横断基盤として独立させるのが最も自然である。

### 2. foundation は transport / observability だけを持つ
- Decision:
  foundation 配下の `progress` は notifier と event payload transport のみを提供し、何を何%と解釈するかは `workflow` が持つ。`telemetry` も logger / context / span / Wails bridge のみを保持し、ユースケース進行判断は持たない。
- Rationale:
  横断基盤に業務進行の意味づけを持ち込むと、別の責務区分が foundation を通じて暗黙結合する。

### 3. LLM 周辺だけを先行移行する
- Decision:
  最初の移行対象は `pkg/gateway/llm/**` と、その依存先になる LLM 向け runtime / workflow / controller だけに限定する。
- Rationale:
  既存の gateway 境界違反の主要因が LLM 周辺に集中しており、先にここを片付ける方が change の粒度を保ちやすい。

### 4. `depguard` は foundation を新しい責務区分として扱う
- Decision:
  `.golangci.yml` に foundation 用 files ルールを追加し、各区分から foundation への依存は許可しつつ、foundation から上位区分への逆依存は禁止する。
- Rationale:
  依存方向の例外運用ではなく、責務区分として明文化して検査できる状態にする必要がある。

## Migration Plan

1. `architecture.md` に `foundation` 区分と依存方向を追加する。
2. `backend-quality-gates` と `.golangci.yml` を更新し、foundation ルールを追加する。
3. `telemetry` と `progress` を foundation 配下へ移し、互換 import の切り替え方針を決める。
4. LLM 周辺の `gateway` / `runtime` / `workflow` / `controller` から foundation を参照するよう置き換える。
5. `backend:lint:file` と `lint:backend` で境界違反が収束することを確認する。

## Risks / Trade-offs

- [Risk] foundation が何でも置ける shared 置き場になる
  - Mitigation: telemetry と progress のような横断基盤に限定し、architecture に責務制限を書く。
- [Risk] package 移動で Wails wiring や logger provider が壊れる
  - Mitigation: LLM 周辺の import だけを先に切り替え、provider graph を lint と起動確認で確認する。
- [Risk] progress の意味づけが foundation 側に漏れる
  - Mitigation: task / design に「progress 解釈は workflow が保持する」と明記する。
