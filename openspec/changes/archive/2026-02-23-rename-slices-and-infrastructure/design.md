# Design: Rename Slices and Infrastructure

## Background
現在のプロジェクト構成では、一部のコンポーネント名が冗長（`export-slice`）であったり、機能的に密結合なインフラストラクチャ層が細分化（`llm-client`, `llm-manager`）されていたりします。また、`database` や `job_queue` といった汎用的な名称を、Go の慣習やプロジェクトの将来的な拡張性に合わせて `datastore`, `queue`, `telemetry` などに洗練させます。

## Goals
- シンプルかつ一貫性のある命名体系への移行。
- LLM 関連コンポーネントの統合によるパッケージ構成の簡素化。
- インフラストラクチャ層のドキュメント不足を解消し、設計意図を明確化する。

## Decisions

### 1. Renaming Strategy
- `specs/` および `pkg/` のディレクトリ名を一斉に変更する。
- Go のパッケージ名もディレクトリ名に合わせて変更する。
- `import` パスおよび `specs` 内の相対リンクをすべて検索・置換で更新する。

### 2. Integration of LLM Components
- `pkg/infrastructure/llm_client` と `pkg/infrastructure/llm_manager` を `pkg/infrastructure/llm` に統合する。
- `llm_manager` の機能は `llm` パッケージの `manager.go` として配置する。
- `llm_client` の各プロバイダ（gemini, xai, local）は `llm/gemini` のようにサブパッケージとして維持、または平滑化を検討する（今回はサブパッケージ維持）。

### 3. Missing Infrastructure Specs
以下のディレクトリを `specs/` に作成し、基本構造（`spec.md`, `design.md` 等）を配置する。
- `specs/datastore/`
- `specs/queue/`
- `specs/telemetry/`
- `specs/progress/`

## Risks / Trade-offs
- **破壊的変更**: 大規模な命名変更はすべてのブランチに影響するため、マージタイミングの調整が必要。
- **リファクタリングの連鎖**: 命名変更に伴い、インターフェース名やメソッド名の変更も誘発される可能性があるが、今回はディレクトリ/パッケージ名の変更を主軸とし、内部シンボルは必要最小限に留める。
- **検索・置換の漏れ**: 文字列リテラルやコメント内での参照が漏れる可能性がある。
