## Why

現在の `pkg/workflow/**` には `gateway` や `controller` への直接依存が残っており、`architecture.md` が定義する orchestration 専用の責務と一致していない。workflow が外部 I/O や UI 境界の詳細を抱えたままだと、slice と runtime の分離も崩れ、artifact を介した受け渡し原則が徹底できない。

## What Changes

- workflow は `controller`、`artifact`、`gateway` を直接 import せず、`slice` と `runtime` を束ねる orchestration に集中する方針を明文化する。
- workflow が現在直接持っている gateway 依存や UI 境界依存を、runtime 契約または controller 側境界へ移す移行方針を定義する。
- slice 間受け渡しに必要な共有データは workflow 自身が保持せず、artifact 識別子や検索条件だけを束ねる方針を明記する。
- `depguard` を workflow 境界に整合するよう更新し、`pkg/workflow/**` の本番コードとテストコードで違反を検出できるようにする。
- 既存 workflow 実装を、更新後の orchestration 境界に沿って段階的に修正する。

## Capabilities

### New Capabilities
- `workflow-orchestration-boundary`: workflow が slice と runtime を束ねる orchestration に専念し、gateway や UI 境界の詳細を持たないことを扱う

### Modified Capabilities
- `backend-quality-gates`: workflow 境界違反を `depguard` で検出する要件へ更新する

## Impact

- 影響範囲は `pkg/workflow/**`、関連する `pkg/runtime/**`、`pkg/controller/**`、`pkg/artifact/**`、および `.golangci.yml` の依存方向 lint 設定。
- workflow から直接参照している gateway、controller、artifact 実装は、runtime 契約、controller 境界、artifact 識別子管理へ移し替えが必要になる。
- test code も含めて workflow を orchestration 専用へ寄せるため、既存の統合テスト配置や依存関係の見直しが必要になる。
