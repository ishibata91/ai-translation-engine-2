<artifact id="proposal" change="fix-spec-deviations" schema="spec-driven">

## Why

現在の実装において、最新のフロントエンド/バックエンド構想（`frontend_architecture.md`, `backend_coding_standards.md` 等）や要求仕様と実際のコード間にいくつかの乖離（Deviations）や未完了タスクが発生しています。
各コーディング標準（コンテキスト伝播、ロギング、VSA 境界など）の徹底が十分でないことが挙げられます。
本チェンジではこれらの仕様と実装の乖離を網羅的に洗い出し、システム全体が本来の仕様要請（Interface-First AIDD / VSA / Schema-Driven）に完全に整合するよう修正を行うことを目的とします。

## What Changes

- **バックエンド規約への準拠 (Backend Standards Compliance)**
  - `pkg/` 以下のすべてのパッケージにおける `context.Context` の適切な伝播と `slog.*Context` による構造化ログの徹底。
  - WET 原則および VSA に反する「安易な共通 DTO 利用」や「不適切な依存関係」がないかの静的解析・修正（`backend:lint` の厳格化）。
- **フロントエンド規約への準拠 (Frontend Standards Compliance)**
  - `pages` と `hooks/features/` 間の依存境界の再点検。DOM構築責務とロジック責務が完全に分離されているかを検証。
  - `any` の駆除、および境界値の `unknown` ＋ `zod` または `valibot` パースの徹底。
- **アーティファクトとソースコードの整合性チェック**
  - 現在のドキュメントに記載されている各パッケージ・モジュールの責務と、実際のコードが持つ責務の同一性を確保。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- なし (仕様に基づく実装の乖離解消のみであり、仕様自体の要件変更はありません)

## Impact

- `pkg/` 以下の各スライスにおける関数シグネチャ（コンテキスト伝播漏れがあった場合の影響）およびログ出力箇所
- フロントエンド `src/` 配下の型定義、依存およびバリデーションロジック部分

</artifact>
