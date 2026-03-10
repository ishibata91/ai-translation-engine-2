## Context

現在の `workflow/config` には contract、SQLiteStore、migration、TypedAccessor、master persona prompt default が混在している。加えて `pkg/gateway/config` は `workflow/config` の単純エイリアスになっており、gateway が workflow を import する構造違反を隠したまま固定化している。

この change では、LLM 周辺だけを対象に config 関連責務を 3 つへ分割する。保存・migration は `gateway/configstore`、実行時読取補助は `runtime/configaccess`、persona slice に関わる default / 解釈は `workflow/persona` に分け、slice 中心で責務を追える形にした上で gateway の再エクスポート構造を廃止する。

## Goals / Non-Goals

**Goals:**
- `pkg/gateway/configstore` に store contract、SQLite 実装、migration を寄せる。
- `pkg/runtime/configaccess` に TypedAccessor と実行時読取補助を寄せる。
- `pkg/workflow/persona` に persona slice 固有の default / 解釈を寄せる。
- LLM 周辺で `gateway -> workflow` 依存を解消する。

**Non-Goals:**
- repository 全体の config import を一括で移行しない。
- config UI や Wails binding の振る舞いを全面改修しない。
- foundation 追加や telemetry / progress 移設をこの change に混ぜない。

## Decisions

### 1. 永続化 contract と migration は `gateway/configstore` に置く
- Decision:
  `Config`、`UIStateStore`、`SecretStore`、`SQLiteStore`、migration は gateway 側の外部保存 adapter として `configstore` に配置する。
- Rationale:
  設定保存は外部資源との技術接続であり、workflow 固有責務ではない。

### 2. TypedAccessor は `runtime/configaccess` に置く
- Decision:
  provider、model、endpoint、concurrency のような実行時設定を読む補助は runtime 側へ寄せる。
- Rationale:
  値の保存ではなく、実行時に都合のよい形へ読み出す責務だからである。

### 3. persona 向け prompt default は `workflow/persona` に寄せる
- Decision:
  master persona prompt の default 値や merge は `pkg/workflow/persona` に置き、persona slice に対応する workflow 側解釈として分離する。
- Rationale:
  これは保存技術ではなく persona slice を進めるための workflow 解釈であり、gateway や runtime に置くべきではない。

### 4. スコープは LLM 周辺だけに限定する
- Decision:
  先行移行対象は LLM manager、model catalog、persona prompt、queue worker など LLM 関連の設定利用箇所に限定する。
- Rationale:
  config 境界の再編を安全に進めるには、利用箇所が集中している LLM 周辺から切り出すのが妥当である。

## Migration Plan

1. `workflow/config` の中身を store / access / `workflow/persona` に分類する。
2. `gateway/configstore`、`runtime/configaccess`、`workflow/persona` の package 形を定義する。
3. LLM 周辺の import と provider wiring を新しい package へ切り替える。
4. `pkg/gateway/config` の再エクスポート構造を除去する。
5. lint と LLM 周辺テストで回帰がないことを確認する。

## Risks / Trade-offs

- [Risk] package rename が広がり過ぎる
  - Mitigation: LLM 周辺だけを対象にし、他機能は後続 change に分離する。
- [Risk] migration / store contract の移動で Wire graph が壊れる
  - Mitigation: provider 単位で lint:file を回し、最後に backend lint で検証する。
- [Risk] persona 向け prompt default が runtime 側へ漏れる
  - Mitigation: persona slice に関わる default / 解釈は `workflow/persona` に閉じ込める。
