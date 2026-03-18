# AI Assistant ルール定義

このファイルはこのプロジェクト向けの常設ルールです。毎回の指示を省くため、以下に従ってください。

## 出力言語
- 返答、資料、プランはすべて日本語で出力する。

## 実行ポリシー
- CLI 実行は削除系以外で確認を求めない。
- 削除・破壊的操作は必ず事前確認する。

## 進め方
- go用に `go-llm-lens` MCPが入っている｡コードの初動探索時は基本的にこれを使うこと｡
- `server-filesystem` MCPが入っている｡検索､書き込み､読み取りは"""必ず"""これを利用すること｡
- 書き込みも原則`server-filesystem`を利用すること｡
- [docs/governance/architecture/spec.md](docs/governance/architecture/spec.md) を常に参照して実装・提案を行う。
- 文書の責務境界や spec の置き場を整理・提案する場合は [docs/governance/spec-structure/spec.md](docs/governance/spec-structure/spec.md) を常に参照する。
- 仕様差分は `changes/<id>/` に置き、正本は `docs/` として扱う。
- `changes/<id>/` では少なくとも `ui.md` `scenarios.md` `logic.md` `tasks.md` を管理対象候補として扱う。
- `docs/` へ同期するのは `ui` と `scenario` レベルの仕様までとし、`logic` の実装詳細は同期しない。
- 設計作業では `skills/` 配下の skill を使い分ける。
- どの設計 skill から始めるか迷う場合は `aite2-design-orchestrator` を入口にする。
- UI 設計は `aite2-ui-design`、シナリオ設計は `aite2-scenario-design`、ロジック設計は `aite2-logic-design` を使う。
- 既存 UI の見た目修正は `aite2-design-fix`、仕様差分の docs 同期は `aite2-sync-docs` を使う。
- 全ての skill で、対話内でタスク化し 1 ステップずつ進める。
- バックエンドを触る場合は、[docs/governance/backend-coding-standards/guide.md](docs/governance/backend-coding-standards/guide.md) を常に参照して実装・提案を行う。
- バックエンドの品質ゲート方針は [docs/governance/backend-quality-gates/spec.md](docs/governance/backend-quality-gates/spec.md) を参照する。
- テスト設計やテスト仕様の検討では [docs/governance/standard-test/spec.md](docs/governance/standard-test/spec.md) を参照する。
- ログ設計やログ運用の検討では [docs/governance/log-guide/spec.md](docs/governance/log-guide/spec.md) を参照する。
- バックエンドを触る場合は、変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行し、その結果を確認してから修正を進める。
- バックエンドの変更では、AI は `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進める。
- フロントを触る場合は、[docs/governance/frontend-architecture/spec.md](docs/frontend/frontend-architecture/spec.md) を常に参照して実装・提案を行う。
- フロントを触る場合は、[docs/governance/frontend-coding-standards/spec.md](docs/frontend/frontend-coding-standards/spec.md) を常に参照して実装・提案を行う。
- フロントを触る場合は、変更中ファイルに対して `npm run lint:file -- <file...>` を逐次実行し、その結果を確認してから修正を進める。
- フロントの変更では、AI は `lint:file -> 修正 -> 再実行 -> typecheck -> lint:frontend -> Playwright` の順で進める。
- フロントの変更を行ったら、作業完了前に `npm run typecheck`、`npm run lint:frontend`、Playwright E2E を必ず実行する。
- フロントのlintはfrontend/に入らないと実行出来ない｡

## バックエンド品質ルール
- バックエンド修正を行ったら、作業中または完了前に必ず `npm run lint:backend` を実行する。
- バックエンド修正では、必要に応じて `npm run backend:lint:file -- <file...>` を先行実行し、ファイル単位で lint を潰しながら進める。
- 必要に応じて `npm run backend:check` や `npm run backend:watch` を使い、ローカルで品質確認を回す。

## 設計・提案
- 設計時に必要なライブラリの提案を行う。
- デファクトスタンダードと言えないライブラリは提案・採用しない。
- UI 設計では Mermaid の state machine を用いる。
- シナリオ設計では Mermaid のロバストネス図を必須とする。
- ロジック設計は 1 主シナリオ単位の責務分解に留め、広げすぎない。

## Wails 開発環境

このプロジェクトは [Wails v2](https://wails.io/) を使用してGoバックエンドとReactフロントエンドを単一の実行ファイルにバンドルします。

### セットアップ確認

```powershell
wails doctor
```

依存関係（WebView2、Node.js、npm、cgo環境）に問題がないことを確認する。

### 開発モード

```powershell
wails dev
```

Wailsウィンドウとホットリロード付きVite devサーバーが起動する。`frontend/src/wailsjs/` は初回起動時に自動生成される。

### 本番ビルド

```powershell
wails build
```

`build/bin/ai-translation-engine-2.exe` が生成される。単体で起動可能な実行ファイル。

### バックエンドテスト（Wails独立）

```powershell
go test ./pkg/...
```

Wailsとは独立してバックエンドのユニットテストを実行できる。
