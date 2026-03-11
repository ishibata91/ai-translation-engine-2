## Why

現在の `pkg/slice/**` には `runtime`、`gateway`、`workflow`、他 slice への直接依存が残っており、`openspec/specs/architecture.md` が定義する責務境界と一致していない。加えて、slice ごとのローカル永続化と slice 間共有データの置き場が曖昧なため、依存方向 lint を強化しても移行先の判断基準が不足している。

## What Changes

- slice は `artifact` 以外の他区分へ直接依存してはならないことを、アーキテクチャ方針として明確化する。
- slice ごとの SQLite などローカル永続化は各 slice 内に保持し、他 slice から参照される共有データのみ `artifact` に置く方針を明記する。
- slice 間受け渡しは `workflow` が `artifact` 識別子や検索条件を束ねて行い、slice 直接依存を禁止する。
- `depguard` を上記方針に整合するよう更新し、`pkg/slice/**` の本番コードとテストコードで境界違反を検出できるようにする。
- 既存の slice 実装とテストを、更新後の境界ルールに適合する形へ段階的に移行する。

## Capabilities

### New Capabilities
- `slice-local-persistence-and-artifact-boundary`: slice ローカル永続化と slice 間共有データの責務境界を定義し、共有データは `artifact` に集約することを扱う

### Modified Capabilities
- `backend-quality-gates`: `depguard` により slice 境界違反を `architecture.md` 準拠で検出する要件へ更新する

## Impact

- 影響範囲は `pkg/slice/**`、関連する `pkg/workflow/**`、`pkg/artifact/**`、および `.golangci.yml` の依存方向 lint 設定。
- slice から直接参照している `runtime`、`gateway`、`workflow`、他 slice の contract / DTO / helper は、artifact 経由または workflow 経由へ移行が必要になる。
- SQLite など slice 固有 DB の配置方針を明文化するが、共有データを `artifact` へ移す場合は保存先設計と既存データの扱いを確認する必要がある。
