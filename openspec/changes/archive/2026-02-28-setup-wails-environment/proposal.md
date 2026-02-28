# Proposal: Go Wails 環境構築

## Problem & Motivation

現在のアーキテクチャでは、バックエンド（`/pkg` 配下のGoコード）とフロントエンド（`mocks-react/`）が完全に分離されており、エンドユーザーが利用するには以下の手順が必要になる。

> **注記**: 現時点の `mocks-react/` はモックとして作成されたものであり、Wails本番フロントエンドへの完全移植は行わない。フロントエンドの本格的な実装はmockの再設計後に別チェンジとして行う。

- Goサーバーの起動
- Node.js / npm による React アプリのビルド・配信
- 複数プロセスの管理

この構成はエンドユーザーには複雑すぎる。本ツールの配布形態として **「単一のexeファイルをダブルクリックするだけで動く」** ネイティブデスクトップアプリを実現したい。

## Proposal

[Wails v2](https://wails.io/) を利用して、GoバックエンドとReactフロントエンドを**単一のネイティブバイナリ（.exe）にバンドル**する環境を構築する。

具体的な方針：
- 既存の `/pkg` パッケージ群をWailsのバックエンド（`App` 構造体）からバインドして呼び出す
- Wailsプロジェクトはリポジトリルート直下に初期化する（`wails.json` を配置）
- ビルドコマンド `wails build` を実行するだけで配布用 `.exe` が生成される構成にする
- **フロントエンドは最小限のプレースホルダー（動作確認用の単一ページ）のみ配置する**。`mocks-react/` の完全移植は行わない

## Capabilities

### New Capabilities

- `wails-app-shell`: Wailsアプリのエントリポイント（`main.go`、`app.go`）を提供するシェル。`/pkg` 配下のドメインロジックをWailsのバインディング経由でフロントエンドに公開する。
- `wails-frontend-placeholder`: Wails動作確認用の最小フロントエンド。`frontend/` ディレクトリに Vite + React の骨格のみ配置し、Wailsランタイム（`@wailsapp/runtime`）との接続を確認できる状態にする。本格的なUIの実装は行わない。

## Impact

- リポジトリルートに `wails.json`、および Wails 用の `main.go`（または既存 `main.go` の更新）を追加する
- `go.mod` に `github.com/wailsapp/wails/v2` を追加する
- `frontend/` ディレクトリに Vite + React のプレースホルダーを新規作成する（`mocks-react/` には一切手を加えない）
- 既存の `cmd/` 配下のCLIエントリポイントには影響を与えない（Wailsは独立したビルドターゲットとして扱う）
- 開発時は `wails dev` コマンドでホットリロード付きの開発環境が利用可能になる
- `mocks-react/` は引き続き独立したモックとして維持し、UIの設計・検討に使用する
