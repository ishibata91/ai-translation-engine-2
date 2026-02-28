# Design: Go Wails 環境構築

## Context

既存プロジェクトの構成:
- **モジュール**: `github.com/ishibata91/ai-translation-engine-2`
- **バックエンド**: `/pkg` 配下にVSA（Vertical Slice Architecture）で構築されたGoコード
  - `pkg/pipeline`: パイプライン管理（`Slice` インターフェース、`Manager`）
  - `pkg/infrastructure`: LLM、DB（SQLite）、Queue、Progress、Telemetry
  - `pkg/translator`, `pkg/terminology`, `pkg/persona`, `pkg/summary`, `pkg/dictionary`, `pkg/export`, `pkg/config`
- **DI**: `github.com/google/wire` によるプロバイダーセット
- **フロントエンドモック**: `mocks-react/` (Vite + React + Tailwind) — Wailsには移植しない
- **既存エントリポイント**: `cmd/` 配下（CLIツール群） — 変更しない

---

## Architecture Decision

### Wails の配置方針

```
ai-translation-engine-2/
├── main.go          ← Wails エントリポイント（新規追加）
├── app.go           ← Wailsバインディング App 構造体（新規追加）
├── wails.json       ← Wails プロジェクト設定（新規追加）
├── build/           ← Wailsビルド成果物ディレクトリ（新規追加）
│   └── appicon.png
├── frontend/        ← Wailsフロントエンド（最小プレースホルダー）（新規追加）
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.tsx
│       └── App.tsx   ← 動作確認用の最小ページ
├── pkg/             ← 既存。変更なし
├── cmd/             ← 既存。変更なし
└── mocks-react/     ← 既存モック。変更なし
```

### なぜルートに `main.go` を置くか

Wailsは慣例として `main.go` をプロジェクトルートに置く。既存の `cmd/` はサブコマンド用CLIであるため独立しており、競合しない。ただし、`go build ./...` によって誤ってWailsビルドが走らないよう、**`//go:build ignore` タグを検討しない**（Wailsは独自のビルドシステムを持つため問題なし）。

---

## App 構造体の設計（`app.go`）

`App` 構造体はWailsのバインディングレイヤーとして機能し、フロントエンドから呼び出せるメソッドを公開する。

**このフェーズでは最小限の実装のみとする**:
- `Greet(name string) string` — 動作確認用のサンプルメソッド（Wails標準テンプレートのまま）
- `/pkg` 配下の実際のドメインロジックのバインディングは **将来の別チェンジ** で実装する

```go
// app.go（概念）
type App struct {
    ctx context.Context
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) Greet(name string) string { return "Hello, " + name }
```

### DI（Wire）との関係

今フェーズでは Wire は使用しない。`App` 構造体は直接インスタンス化する。将来的に `/pkg` の機能をバインドする際に Wire の Provider を `App` のコンストラクタへ注入する設計に移行する。

---

## フロントエンド（`frontend/`）の設計

**方針**: `wails init` が生成するデフォルトの Vite + React テンプレートを**そのまま使用**し、改変を最小限にする。

- `@wailsapp/runtime` のインポートと `window.go.main.App.Greet()` 呼び出しの動作確認のみを行う
- `mocks-react/` とは完全に独立したディレクトリとして管理する
- CSS/UIの実装は行わない（後の本格フロントエンドチェンジで行う）

---

## `wails.json` の主要設定

```json
{
  "name": "ai-translation-engine-2",
  "outputfilename": "ai-translation-engine-2",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "wailsjsdir": "./frontend/src/wailsjs"
}
```

- `wailsjsdir`: Wailsが自動生成するTypeScriptバインディングの出力先
- `frontend:dev:serverUrl: "auto"`: `wails dev` 実行時にViteのdevサーバーを自動起動する

---

## `go.mod` への影響

```
require (
    github.com/wailsapp/wails/v2 v2.x.x  // 追加
    ...既存の依存関係...
)
```

Wailsはcgoを必要とする（Windows向けWebView2レンダラーのため）。`mattn/go-sqlite3` も既にcgoを使用しているため、環境的に問題はない。

---

## ビルドフロー

| コマンド            | 用途                                                           |
| ------------------- | -------------------------------------------------------------- |
| `wails dev`         | 開発モード（ホットリロード付き）。Vite devサーバーを内部で起動 |
| `wails build`       | 本番ビルド。`build/bin/ai-translation-engine-2.exe` を生成     |
| `go test ./pkg/...` | バックエンドのユニットテスト（Wailsと独立して実行可能）        |

---

## Risks / Trade-offs

| リスク                                                                                                            | 対策                                                                                      |
| ----------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| **WebView2の依存**: Windows環境ではWebView2ランタイムが必要。未インストール環境では実行不可                       | Wails v2はWebView2インストーラーを同梱するオプション（`-webview2 embed`）がある。将来対応 |
| **cgo環境**: WindowsでのGoビルドにはC コンパイラ（MinGW等）が必要                                                 | 既に `mattn/go-sqlite3` でcgoを使用しているためビルド環境は既に整っているはず             |
| **`wails init` との競合**: `wails init` はルートに `main.go` を生成するが、既存プロジェクト構造との競合可能性あり | `wails init` を使わず手動で必要ファイルを作成するか、テンプレートを確認してから適用する   |
