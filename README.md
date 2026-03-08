# AI Translation Engine 2 (WIP)

> **Status:** Work in Progress（開発途中）

Skyrim 向け翻訳ワークフローを支援する、Wails v2 ベースのデスクトップアプリです。  
Go バックエンドと React フロントエンドを単一バイナリに統合し、翻訳作業の前処理・辞書構築・ペルソナ生成を支援します。

## 主な機能（現状）

- ダッシュボードでジョブ進捗を確認・再開・停止
- 辞書ビルダーで SSTXML などの辞書データをインポート・編集
- マスターペルソナ生成（JSON 入力 + LLM 設定）

※ 一部画面・機能は未実装または暫定実装です。

## 技術スタック

- Desktop: Wails v2
- Backend: Go
- Frontend: React + TypeScript + Vite + Zustand + Tailwind CSS + daisyUI
- Quality Gate:
  - Backend: `golangci-lint` / `goimports` / `goleak` など
  - Frontend: ESLint / TypeScript / Vitest

## セットアップ

前提:

- Go
- Node.js / npm
- Wails v2

```bash
# ルート依存（backend quality scripts など）
npm install

# フロント依存
npm --prefix frontend install
```

## 開発コマンド

### Wails 開発モード

```bash
wails dev
```

### フロントエンド

```bash
npm --prefix frontend run dev
npm --prefix frontend run lint
npm --prefix frontend run test
npm --prefix frontend run build
```

### バックエンド品質チェック

```bash
npm run backend:lint
npm run backend:test
npm run backend:check
```

## 注意事項（WIP）

- `frontend/src/wailsjs/` は Wails 実行時に生成されるファイルを含みます。
- テスト/ビルド時に Wails バインディングが必要なケースでは、先に `wails dev` または `wails build` を実行してください。
- 現在、設計・実装ともに継続的に更新中です。破壊的変更が入る可能性があります。

## 今後の予定（例）

- 未実装ページの具体化
- テストの安定化（Wails 生成物依存の整理）
- README / アーキテクチャ図 / 開発フローの更新

---

本 README は暫定版です。必要に応じて随時更新します。
