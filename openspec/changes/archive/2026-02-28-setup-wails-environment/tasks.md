# Tasks: Go Wails 環境構築

## 1. 事前準備・環境確認

- [x] 1.1 Wails CLI がインストールされているか確認する（`wails version`）
- [x] 1.2 未インストールの場合は `go install github.com/wailsapp/wails/v2/cmd/wails@latest` でインストールする
- [x] 1.3 `wails doctor` を実行し、依存関係（WebView2、Node.js、npm、cgo環境）に問題がないか確認する
- [x] 1.4 既存の `go.mod` の Goバージョンを確認し、Wails v2 の要件（Go 1.18以上）を満たしているか確認する

---

## 2. Wailsプロジェクトのセットアップ（`wails.json` + Goエントリポイント）

- [x] 2.1 リポジトリルートに `wails.json` を手動で作成する（`wails init` は使用しない）
  - `name`, `outputfilename`, `frontend:install`, `frontend:build`, `frontend:dev:watcher`, `frontend:dev:serverUrl`, `wailsjsdir` を設定する
  - `wailsjsdir` は `"./frontend/src/wailsjs"` に設定する
- [x] 2.2 `go get github.com/wailsapp/wails/v2` を実行し、`go.mod` / `go.sum` を更新する
- [x] 2.3 リポジトリルートに `main.go` を作成する
  - `package main` / `//go:build !ignore` でビルド条件を確認する
  - `wails.Run()` を呼び出し、`App` インスタンス・ウィンドウ設定（タイトル `"AI Translation Engine"`、幅 `1280`、高さ `800`）・`frontend/dist` の `embed.FS` を渡す
- [x] 2.4 リポジトリルートに `app.go` を作成する
  - `App` 構造体（`ctx context.Context` フィールド）を定義する
  - `NewApp() *App` コンストラクタを実装する
  - `startup(ctx context.Context)` ライフサイクルフックを実装する
  - `shutdown(ctx context.Context)` ライフサイクルフックを実装する（このフェーズでは空でよい）
  - `Greet(name string) string` バインドメソッドを実装する（`"Hello, " + name` を返す）

---

## 3. Wailsフロントエンドプレースホルダーの作成（`frontend/`）

- [x] 3.1 `frontend/` ディレクトリを作成する（`mocks-react/` には一切手を加えない）
- [x] 3.2 `frontend/package.json` を作成する
  - `"name"`: `"frontend"`
  - `"scripts"`: `"dev": "vite"`, `"build": "vite build"`, `"preview": "vite preview"`
  - `devDependencies`: `react`, `react-dom`, `vite`, `@vitejs/plugin-react`, TypeScript関連
- [x] 3.3 `frontend/vite.config.ts` を作成する（`@vitejs/plugin-react` を使用するシンプルな設定）
- [x] 3.4 `frontend/src/main.tsx` を作成する（ReactDOM.createRoot で App をマウントする）
- [x] 3.5 `frontend/index.html` を作成する（Viteのエントリhtmlとして `src/main.tsx` をスクリプトソースに指定）
- [x] 3.6 `frontend/` ディレクトリで `npm install` を実行し、依存関係をインストールする
- [x] 3.7 `frontend/src/wailsjs/` ディレクトリを `.gitignore` に追加する（Wailsが自動生成するため）

---

## 4. 動作確認用UIの実装（`frontend/src/App.tsx`）

- [x] 4.1 `frontend/src/App.tsx` に動作確認用コンポーネントを実装する
  - 名前入力テキストフィールド（state管理）
  - 「Greet」ボタン（クリック時に `Greet(name)` を呼び出す）
  - 結果表示エリア（Goから返された文字列を表示）
- [x] 4.2 `frontend/src/wailsjs/go/main/App.js`（Wailsが生成する型定義）を呼び出す形でGoバインドを呼び出す
  - ※ `wails dev` 初回起動時に `wailsjs/` が自動生成されるため、まず `wails dev` を実行してから参照する

---

## 5. 動作確認

- [x] 5.1 `wails dev` を実行し、Wailsウィンドウが起動することを確認する
- [x] 5.2 フロントエンドの動作確認ページが表示されることを確認する
- [x] 5.3 テキストフィールドに名前を入力してGreetボタンをクリックし、`"Hello, <name>"` が表示されることを確認する（GoバインドがJSから正常に呼び出されていることの確認）
- [x] 5.4 `wails build` を実行し、`build/bin/ai-translation-engine-2.exe` が生成されることを確認する
- [x] 5.5 生成された `.exe` をダブルクリックして、単体で起動することを確認する（Node.js・Goランタイム不要）

---

## 6. 後片付け・確認

- [x] 6.1 `go test ./pkg/...` を実行し、既存バックエンドのテストが引き続き通ることを確認する（Wails追加による回帰がないことの確認）
- [x] 6.2 `.gitignore` に `build/bin/`, `frontend/node_modules/`, `frontend/dist/`, `frontend/src/wailsjs/` を追加する
- [x] 6.3 `AGENTS.md` または README に Wails 開発環境のセットアップ手順（`wails doctor` / `wails dev` / `wails build`）を追記する
