# Wails アプリシェル

## Overview

GoバックエンドとWailsランタイムをつなぐアプリケーションシェル。
`main.go` と `app.go` を中心に、Wailsアプリのライフサイクル管理、controller の依存注入、フロントエンドへの Go API バインディングを担う。

Wails が公開する Go API は `pkg/controller` 配下の controller に限定し、`main.App` や usecase slice / service を直接 bind しない。`App` はライフサイクル保持専用のホストとし、UI 向け API は controller 経由で公開する。

---

## Files

| ファイル  | 役割                                                                  |
| --------- | --------------------------------------------------------------------- |
| `main.go` | Wailsアプリのエントリポイント。controller の組み立て、`wails.Run()`、`Bind` と `OnStartup` を担う |
| `app.go`  | `App` 構造体の定義。Wailsライフサイクルフックと runtime context 保持を担う |

---

## Requirements

### Requirement: Wailsアプリの起動

- **GIVEN** ユーザーが `.exe` をダブルクリックする
- **WHEN** `main()` が実行される
- **THEN** Wailsウィンドウが起動し、`frontend/` のReactアプリが表示される

### Requirement: ライフサイクルフック

- **GIVEN** Wailsアプリが起動する
- **WHEN** `startup(ctx)` が呼ばれる
- **THEN** `App` 構造体が `context.Context` を保持し、Wails runtime を利用する controller へ共有できる状態にならなければならない
- **AND** `OnStartup` では bind 済み controller のうち `context.Context` を必要とするものへ runtime context を注入しなければならない

- **GIVEN** Wailsアプリが終了する
- **WHEN** `shutdown(ctx)` が呼ばれる
- **THEN** 保持したリソースを安全にクリーンアップする

### Requirement: Wails公開APIはcontroller配下に限定されなければならない

Wails がフロントエンドへ公開する Go API は `pkg/controller` 配下の controller 型に限定されなければならない。`main` パッケージの `App` や usecase slice / service を `Bind` に直接含めてはならない。

#### Scenario: main.go の Bind が controller のみを公開する
- **WHEN** `main.go` の Wails `Bind` 設定を確認する
- **THEN** `Bind` に含まれる型は `pkg/controller` 配下の controller のみでなければならない
- **AND** `main.App` や `pkg/modelcatalog.ModelCatalogService`、`pkg/persona.Service` のような controller 外の型を直接 bind してはならない

#### Scenario: UI向けAPIがcontroller経由で公開される
- **WHEN** 辞書 API、モデル一覧 API、persona 読み取り API などの UI 向け機能を Wails へ公開する
- **THEN** それらの公開メソッドは `pkg/controller` 配下の adapter として実装されなければならない
- **AND** controller は入力整形と委譲に限定され、業務ロジック本体を持ってはならない

### Requirement: バインディング変更後はfrontendが新controller APIへ追従しなければならない

Wails の bind 名や生成物が変更された場合、frontend は旧公開名との互換層を残さず、新しい controller ベースの API 参照へ全面追従しなければならない。UI は `frontend_architecture.md` に従い、feature hook 経由で新しい `wailsjs` 生成物を利用しなければならない。

#### Scenario: wailsjs 生成物の参照先をcontrollerへ切り替える
- **WHEN** Wails bind 変更後に `frontend/src/wailsjs` を再生成する
- **THEN** frontend の Wails 呼び出しは新しい controller 名に対応した生成物を参照しなければならない
- **AND** 旧 `main/App` や slice service 名の生成物参照を残してはならない

#### Scenario: pages はcontroller化後もfeature hook経由でWailsを利用する
- **WHEN** frontend が新しい Wails API を利用する
- **THEN** `pages` は `wailsjs` を直接 import せず、`hooks/features/*` 経由で呼び出さなければならない
- **AND** bind 名変更に伴う修正は page ではなく feature hook / adapter に閉じ込めなければならない

---

## Interface

```go
// main.go
package main

// controller を組み立てて wails.Run() で起動する
// Bind には pkg/controller 配下の型のみを渡す
// OnStartup で runtime context を必要な controller へ注入する

// app.go
package main

type App struct {
    ctx context.Context
}

func NewApp() *App
func (a *App) startup(ctx context.Context)
func (a *App) shutdown(ctx context.Context)
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

## Non-Requirements（現フェーズでは実装しない）

- `main.App` への UI 向け API の再集約
- usecase slice / service の Wails 直接 bind
- Wire（DI）の導入
- システムトレイアイコン
- 複数ウィンドウ
- ネイティブメニュー
