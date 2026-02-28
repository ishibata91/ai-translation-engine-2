# JetBrains AI Assistant ルール定義

このファイルは JetBrains AI Assistant 向けの常設ルールです。毎回の指示を省くため、以下に従ってください。

## 出力言語
- 返答、資料、プランはすべて日本語で出力する。

## 実行ポリシー
- CLI 実行は削除系以外で確認を求めない。
- 削除・破壊的操作は必ず事前確認する。

## 進め方
- openspecが指定された場合は必ず従う。勝手に実装を始めない。
- 仕様書など資料の図はmermaidで出力する。

## 設計・提案
- 設計時に必要なライブラリの提案を行う。
- デファクトスタンダードと言えないライブラリは提案・採用しない。

## Wails 開発環境

このプロジェクトは [Wails v2](https://wails.io/) を使用してGoバックエンドとReactフロントエンドを単一の実行ファイルにバンドルします。

### セットアップ確認

```bash
wails doctor
```

依存関係（WebView2、Node.js、npm、cgo環境）に問題がないことを確認する。

### 開発モード

```bash
wails dev
```

Wailsウィンドウとホットリロード付きVite devサーバーが起動する。`frontend/src/wailsjs/` は初回起動時に自動生成される。

### 本番ビルド

```bash
wails build
```

`build/bin/ai-translation-engine-2.exe` が生成される。単体で起動可能な実行ファイル。

### バックエンドテスト（Wails独立）

```bash
go test ./pkg/...
```

Wailsとは独立してバックエンドのユニットテストを実行できる。
