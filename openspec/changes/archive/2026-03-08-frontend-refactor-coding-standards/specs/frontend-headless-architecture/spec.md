## ADDED Requirements

### Requirement: Headless Architecture Boundary Rules Are Lint-Enforced
Headless Architecture の依存境界は、レビュー依存ではなく lint により継続検証されなければならない。`pages` は `hooks/features/*` が返す契約経由で描画に専念し、`wailsjs` や `store` への直接依存を機械的に検出できなければならず、この要件を MUST とする。

#### Scenario: Page Imports Wails Directly
- **WHEN** 開発者が `pages` 配下から `wailsjs/go/...` を直接 import して lint を実行する
- **THEN** lint は Headless Architecture 違反として失敗しなければならない

#### Scenario: Page Uses Feature Hook Boundary
- **WHEN** 開発者が `pages` 配下で feature hook だけに依存して描画を実装し lint を実行する
- **THEN** lint は境界ルールを満たしたものとして通過しなければならない
