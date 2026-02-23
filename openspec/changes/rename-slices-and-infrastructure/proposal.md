# Proposal: Slice and Infrastructure Renaming & Specification Completion

## Motivation
プロジェクトの再編に伴い、コンポーネント名の整合性を高め、冗長な構造を排除します。また、現在 `specs` ディレクトリに不足しているインフラストラクチャ層の仕様書を追加し、ドキュメントの網羅性を向上させます。

## Capabilities
- **Rename**: 
  - `export-slice` を `export` に改名 (specs および pkg)
  - `llm-client` と `llm-manager` を `llm` に統合・改名 (pkg/infrastructure)
  - `database` を `datastore` に改名 (pkg/infrastructure)
  - `job_queue` を `queue` に改名 (pkg/infrastructure)
  - `logger` を `telemetry` に改名 (pkg/infrastructure)
- **Specification Completion**: 以下のインフラコンポーネントの仕様書を `specs/` に作成します。
  - `datastore` (旧 database)
  - `queue` (旧 job_queue)
  - `telemetry` (旧 logger)
  - `progress` (進捗管理)

## Use Cases
- 開発者がより直感的で簡潔な名前でコンポーネントを参照できる。
- インフラストラクチャ層の各コンポーネントの役割と仕様が明確化される。

## Success Criteria
- 該当するすべてのディレクトリ・パッケージ名が新しい名称に変更されている。
- `specs/` 以下の古い名称のディレクトリが削除され、新しい名称に移行されている。
- 不足していたインフラストラクチャ仕様書（datastore, queue, telemetry, progress）が `specs/` に追加されている。
- プロジェクト全体の参照（import, specs 内のリンク等）が更新されている。
- ビルドとテストが正常に動作し、命名変更による不具合がないことが確認されている。
