## Why

現在の `pkg/gateway/**` には `workflow` や `runtime` への直接依存が残っており、`architecture.md` が定義する外部資源への依頼口という責務と一致していない。gateway が上位層の進行制御や状態解釈を持つと、runtime からの利用境界が崩れ、外部接続実装の差し替えや検証も難しくなる。

## What Changes

- gateway は外部資源への依頼口に専念し、`controller`、`workflow`、`runtime`、`slice`、`artifact` を直接 import しない方針を明文化する。
- runtime が gateway を利用するための契約境界と、gateway が返す中立的な結果 DTO の方針を定義する。
- gateway が現在保持している workflow / runtime 依存を、純粋な技術接続実装へ寄せる移行方針を定義する。
- `depguard` を gateway 境界に整合するよう更新し、`pkg/gateway/**` の本番コードとテストコードで違反を検出できるようにする。
- 既存 gateway 実装を、更新後の外部接続責務に沿って段階的に修正する。

## Capabilities

### New Capabilities
- `gateway-io-boundary`: gateway が外部資源への依頼口に専念し、上位層の進行制御や状態解釈を持たないことを扱う

### Modified Capabilities
- `backend-quality-gates`: gateway 境界違反を `depguard` で検出する要件へ更新する

## Impact

- 影響範囲は `pkg/gateway/**`、関連する `pkg/runtime/**`、および `.golangci.yml` の依存方向 lint 設定。
- gateway から直接参照している workflow、runtime、slice、artifact 依存は、runtime から消費できる中立契約へ移し替えが必要になる。
- test code も含めて gateway を純粋な技術接続実装へ寄せるため、既存 helper や integration test の責務見直しが必要になる。
