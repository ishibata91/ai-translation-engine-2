# Specs: Rename Slices and Infrastructure

## ADDED Requirements

### Requirement: Consistent Naming
プロジェクト全体での命名規則を統一するため、冗長な「-slice」接尾辞を排除し、インフラストラクチャ層の名称をより適切なものに変更する。

#### Scenario: Rename Export Slice
- **WHEN** ユーザーが `export-slice` を参照する場合
- **THEN** 新しい名称である `export` を使用しなければならない。

#### Scenario: Unified LLM Infrastructure
- **WHEN** LLM 関連の機能（クライアント、マネージャー）を利用する場合
- **THEN** 統合された `llm` パッケージを通じてアクセスしなければならない。

#### Scenario: Standardized Infrastructure Names
- **WHEN** データベース、ジョブキュー、ロガー、進捗管理の機能を参照する場合
- **THEN** それぞれ `datastore`, `queue`, `telemetry`, `progress` という名称を使用しなければならない。

### Requirement: Infrastructure Documentation
すべてのインフラストラクチャコンポーネントは、その役割と仕様を記述したドキュメントを `specs/` 以下に持たなければならない。

#### Scenario: Missing Specs Creation
- **WHEN** `specs/` ディレクトリを確認した際
- **THEN** `datastore`, `queue`, `telemetry`, `progress` の各ディレクトリが存在し、必要な仕様書が含まれていなければならない。

## MODIFIED Requirements

### Requirement: Directory Structure
- `specs/export-slice` -> `specs/export`
- `pkg/export_slice` -> `pkg/export`
- `pkg/infrastructure/llm_client` & `llm_manager` -> `pkg/infrastructure/llm`
- `pkg/infrastructure/database` -> `pkg/infrastructure/datastore`
- `pkg/infrastructure/job_queue` -> `pkg/infrastructure/queue`
- `pkg/infrastructure/logger` -> `pkg/infrastructure/telemetry`
