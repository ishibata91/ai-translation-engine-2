<artifact id="design" change="fix-spec-deviations" schema="spec-driven">

## Context

フロントエンド・バックエンド間の基本設計（`frontend_architecture.md`、`backend_coding_standards.md` 等）は確立していますが、開発を進める中で仕様に基づく実装方針（VSA、Headless UI、コンテキストと構造化ログの伝播）から一部乖離しているコードベースが残存しています。
本設計は、これらの乖離を静的解析とテスト駆動によって一掃するためのプロセスと修正方針を定めます。

## Goals / Non-Goals

**Goals:**
- 対象とするすべてのコードベースが最新の仕様規約ドキュメントに違反しない状態を作ること。
- バックエンドの各パッケージに `slog` と `context.Context` が適切に配線されること。
- フロントエンドの `pages` と `wailsjs` の依存関係を断ち切ること。

**Non-Goals:**
- 本チェンジにおける新規機能の開発や既存機能の仕様変更（外部振る舞いの変更）。
- インフラ構成の変更やミドルウェアの刷新などの大規模リファクタリング。

## Decisions

### 1. バックエンド規約の一律適用手法
- バックエンド全体にわたる `slog` や `context` 漏れの修正は目視ではなくツールに依存します。
- **決定事項**: `backend:lint` および `golangci-lint` をベースにファイル単位で静的解析を行い、警告が出たファイルに対して機械的に修正を加えます。`slog.*Context` を一貫して適用します。

### 2. フロントエンドの責務分離（Headless UI）
- `MasterPersona.tsx` や `DictionaryBuilder.tsx` などの主要ページについては既にリファクタリングが進んでいますが、他のページやコンポーネントにおける `wailsjs` の直接コールが存在しないか確認します。
- **決定事項**: UI コンポーネントで見つかった依存漏れについては、随時 `hooks/features/` 配下へ Hook を切り出します。

## Risks / Trade-offs

- **[Risk]** 広範囲のコード修正により既存のテストが壊れる。
  ➔ *Mitigation*: Go の静的型検査と Table-Driven Test、およびフロントエンドの TypeScript Type Check (`npm run typecheck`) を各段階で通過させながら修正を進めます。

</artifact>
