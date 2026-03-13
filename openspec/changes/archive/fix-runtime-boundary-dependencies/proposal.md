## Why

現在の `pkg/runtime/**` には `workflow` や `slice` への直接依存が残っており、`architecture.md` が定義する実行制御基盤の責務と一致していない。runtime がユースケース進行や slice 固有ロジックを抱えると、workflow の orchestration と外部 I/O 実行の境界が崩れ、gateway 境界も不安定になる。

## What Changes

- runtime は外部 I/O 実行と実行制御基盤に専念し、`workflow`、`slice`、`artifact` を直接 import しない方針を明文化する。
- workflow から runtime へ渡す実行要求と、runtime から workflow へ返す結果の契約境界を整理する。
- runtime が現在保持している workflow / slice 依存を、共通 executor 契約や中立 DTO を用いた境界へ移す方針を定義する。
- `depguard` を runtime 境界に整合するよう更新し、`pkg/runtime/**` の本番コードとテストコードで違反を検出できるようにする。
- 既存 runtime 実装を、更新後の実行制御責務に沿って段階的に修正する。

## Capabilities

### New Capabilities
- `runtime-execution-boundary`: runtime が外部 I/O 実行と実行制御に専念し、workflow / slice 固有知識を持たないことを扱う

### Modified Capabilities
- `backend-quality-gates`: runtime 境界違反を `depguard` で検出する要件へ更新する

## Impact

- 影響範囲は `pkg/runtime/**`、関連する `pkg/workflow/**`、`pkg/gateway/**`、および `.golangci.yml` の依存方向 lint 設定。
- runtime から直接参照している workflow、slice、artifact 依存は、中立契約または返却 DTO 境界へ移し替えが必要になる。
- test code も含めて runtime を実行制御基盤へ寄せるため、既存の integration test や helper 配置の見直しが必要になる。
