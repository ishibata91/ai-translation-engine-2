# Spec: wails-frontend-placeholder

## Overview

Wailsの動作確認用最小フロントエンド。
`frontend/` ディレクトリに Vite + React の骨格を配置し、Wailsランタイムとの接続と、Goバインドメソッド（`Greet`）の呼び出しを確認できる状態を実現する。

**このCapabilityの目的は「Wailsが正常に動作すること」の確認のみ**であり、本格的なUIは実装しない。`mocks-react/` には一切手を加えない。

> **実装上の注意**: フロントエンドは `wailsjs/` への静的インポートではなく、Wailsランタイムが `window.go` に注入するグローバルAPIを直接呼び出す。これにより `wailsjs/` 生成前のビルドエラーを回避する。

---

## Files

| ファイル                  | 役割                                                              |
| ------------------------- | ----------------------------------------------------------------- |
| `frontend/index.html`     | Viteエントリポイント                                              |
| `frontend/package.json`   | npm設定                                                           |
| `frontend/vite.config.ts` | Vite設定                                                          |
| `frontend/tsconfig.json`  | TypeScript設定                                                    |
| `frontend/src/main.tsx`   | Reactアプリのマウントポイント                                     |
| `frontend/src/App.tsx`    | 動作確認用の単一コンポーネント。`window.go` 経由でGoを呼び出す    |
| `frontend/src/wailsjs/`   | Wailsが自動生成するGoバインディング（`wails dev` 初回起動で生成） |

---

## Requirements

### Requirement: フロントエンドのビルド

- **GIVEN** `frontend/` ディレクトリが存在する
- **WHEN** `npm run build`（または `wails build` 経由）を実行する
- **THEN** `frontend/dist/` にビルド成果物が生成され、`main.go` が `embed.FS` で取り込める状態になる

### Requirement: Wailsランタイムとの接続

- **GIVEN** Wailsアプリが起動している
- **WHEN** フロントエンドが読み込まれる
- **THEN** Wailsランタイムが `window.go` にバインディングを注入し、Go側との通信が可能になる

### Requirement: Goバインドメソッドの呼び出し確認

- **GIVEN** フロントエンドの動作確認ページが表示されている
- **WHEN** ユーザーがテキストフィールドに名前を入力してボタンをクリックする
- **THEN** `window.go.main.App.Greet(name)` がGoバックエンドで実行され、戻り値（`"Hello, <name>!"`）がページに表示される

### Requirement: `wails dev` での開発モード動作

- **GIVEN** 開発環境で `wails dev` を実行する
- **WHEN** Vite devサーバーが起動する
- **THEN** ホットリロードが有効になり、フロントエンドの変更が即時反映される

---

## UI 仕様（動作確認用ページのみ）

```
┌─────────────────────────────────────┐
│  AI Translation Engine              │
│                                     │
│  名前: [___________________]        │
│        [Greetボタン]                │
│                                     │
│  結果: Hello, <name>!               │
└─────────────────────────────────────┘
```

- スタイルはシンプルなインラインスタイルのみ使用する
- カスタムデザインシステムは実装しない

---

## 依存関係

| パッケージ             | バージョン | 用途                  |
| ---------------------- | ---------- | --------------------- |
| `react`                | `^18.x`    | UIフレームワーク      |
| `react-dom`            | `^18.x`    | DOMレンダリング       |
| `vite`                 | `^5.x`     | ビルドツール          |
| `@vitejs/plugin-react` | `^4.x`     | ViteのReactプラグイン |
| `typescript`           | `^5.x`     | 型チェック            |

※ Tailwind CSS、Router、状態管理ライブラリは**含めない**（本格実装フェーズで別途導入）

---

## Non-Requirements（現フェーズでは実装しない）

- `mocks-react/` のコードの移植・流用
- ルーティング（React Router等）
- グローバル状態管理（Zustand、Redux等）
- Tailwind CSS等のデザインシステム
- 本番UIページ（ダッシュボード、翻訳フロー等）
