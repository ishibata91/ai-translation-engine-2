# Spec: wails-app-shell

## Overview

GoバックエンドとWailsランタイムをつなぐアプリケーションシェル。
`main.go` と `app.go` の2ファイルで構成され、Wailsアプリのライフサイクル管理とフロントエンドへのバインディングを担う。

このフェーズでは `/pkg` ドメインロジックのバインドは行わず、Wailsの起動・終了ライフサイクルと動作確認用の最小メソッドのみを実装する。

---

## Files

| ファイル  | 役割                                                                  |
| --------- | --------------------------------------------------------------------- |
| `main.go` | Wailsアプリのエントリポイント。`wails.Run()` を呼び出す               |
| `app.go`  | `App` 構造体の定義。Wailsライフサイクルフックとバインドメソッドを持つ |

---

## Requirements

### Requirement: Wailsアプリの起動

- **GIVEN** ユーザーが `.exe` をダブルクリックする
- **WHEN** `main()` が実行される
- **THEN** Wailsウィンドウが起動し、`frontend/` のReactアプリが表示される

### Requirement: ライフサイクルフック

- **GIVEN** Wailsアプリが起動する
- **WHEN** `startup(ctx)` が呼ばれる
- **THEN** `App` 構造体が `context.Context` を保持し、以降のGoメソッド呼び出しで使用できる状態になる

- **GIVEN** Wailsアプリが終了する
- **WHEN** `shutdown(ctx)` が呼ばれる
- **THEN** 保持したリソースを安全にクリーンアップする（このフェーズでは特に行うことはない）

### Requirement: 動作確認用バインドメソッド

- **GIVEN** フロントエンドから `Greet(name)` を呼び出す
- **WHEN** Wailsランタイムがバインディングを経由してGoメソッドを実行する
- **THEN** `"Hello, " + name` の文字列が返され、フロントエンドに表示される

---

## Interface

```go
// main.go
package main

// wails.Run() でAppを起動する
// ウィンドウタイトル、サイズ、Assets（frontend/distを埋め込み）を設定する

// app.go
package main

type App struct {
    ctx context.Context
}

func NewApp() *App
func (a *App) startup(ctx context.Context)
func (a *App) shutdown(ctx context.Context)
func (a *App) Greet(name string) string
```

---

## Wailsウィンドウ設定

| 設定項目       | 値                        |
| -------------- | ------------------------- |
| タイトル       | `"AI Translation Engine"` |
| 幅             | `1280` px                 |
| 高さ           | `800` px                  |
| リサイズ可能   | `true`                    |
| フルスクリーン | `false`                   |
| フレームレス   | `false`                   |

---

## Non-Requirements（このフェーズでは実装しない）

- `/pkg` ドメインロジック（`pipeline`, `translator`, `terminology` 等）のバインド
- Wire（DI）の導入
- システムトレイアイコン
- 複数ウィンドウ
- ネイティブメニュー
