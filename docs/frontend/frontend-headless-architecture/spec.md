# フロントエンド ヘッドレスアーキテクチャ

## Overview

フロントエンドのページを Headless Architecture (Pattern A) に従って分離し、UI 描画と機能ロジックの責務を明確化する。

## Requirements

### Requirement: Frontend Headless Architecture (Pattern A)
The frontend architecture MUST split page-level UI rendering and feature logic using Custom Hooks.

- フロントエンドコンポーネントは、`MasterPersona`、`DictionaryBuilder` などを対象とし、ロジック部分を抽出した Custom Hook (`useMasterPersona`, `useDictionaryBuilder` など) と純粋な UI 描画のみを行うビューコンポーネントに分離しなければならない。
- これら Custom Hook は `src/hooks/features/` 配下に定義され、Wails の API バインディング、`taskStore` や `uiStore` 等の Zustand ストアとの通信、およびローカルな View Model 状態管理をすべて一手に引き受ける。
- ビューコンポーネントは、Custom Hook から返却されるデータとコールバック関数のみを利用して DOM 構造 (JSX) を組み立てる。

#### Scenario: Verify Architecture Split
- **WHEN** Master Persona ページが表示される
- **THEN** ページコンポーネントは状態更新による不要な親要素の再描画を可能な限り抑制しつつ、既存と同等の機能を維持している
- **THEN** Wails (go) への呼び出し処理が JSX ファイル内にハードコードされず、Custom Hook を経由している

### Requirement: Headless Architecture Boundary Rules Are Lint-Enforced
Headless Architecture の依存境界は、レビュー依存ではなく lint により継続検証されなければならない。`pages` は `hooks/features/*` が返す契約経由で描画に専念し、`wailsjs` や `store` への直接依存を機械的に検出できなければならず、この要件を MUST とする。

#### Scenario: Page Imports Wails Directly
- **WHEN** 開発者が `pages` 配下から `wailsjs/go/...` を直接 import して lint を実行する
- **THEN** lint は Headless Architecture 違反として失敗しなければならない

#### Scenario: Page Uses Feature Hook Boundary
- **WHEN** 開発者が `pages` 配下で feature hook だけに依存して描画を実装し lint を実行する
- **THEN** lint は境界ルールを満たしたものとして通過しなければならない
