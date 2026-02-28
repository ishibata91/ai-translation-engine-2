## Why

mocks-reactの静的なUI部品（ダッシュボード、ヘッダー、サイドバー、ログバー、詳細ペイン）のベースが完成し、Wails+Reactの最小構成プロジェクトも立ち上がりました。
次に、これらの独立したモック実装を本番の `frontend` 下に移植し、Wailsのバックエンドと連携するための準備を行う必要があります。アプリの基礎となるレイアウト（コアなナビゲーションやステータス表示領域）とUIフレームワークを早期に統合しておくことで、今後の各画面の実装をスムーズに進められます。

## What Changes

本変更では、モックアプリから本番フロントエンド環境（Vite+React+Tailwind+DaisyUI）へコアUIコンポーネントを移植し、今後の拡張に耐えうる形へ以下の通りリファクタリングを行います。

- `mocks-react` から `frontend` への移植対象:
  - **レイアウト系**: `Layout.tsx`, `Sidebar.tsx`, Header（Layout内から分割）, `DetailPane.tsx`
  - **ログ・ステータス系**: `LogViewer.tsx`, `LogDetail.tsx`
  - **ページ**: `Dashboard.tsx`
- **リファクタリングのポイント**:
  - `Layout.tsx` 内で直書きされているヘッダー部分を `Header.tsx` など独立したコンポーネントとして適切に切り出す。
  - React Router の設定 (`HashRouter` への対応などWails特有のルーティング要件との整合性) を考慮した設計への移行。
  - 横幅をリサイズ可能な `LogViewer` や、下部から展開する `DetailPane` の表示状態（開閉状態・選択中ログ情報）を、上位のLayoutなどローカルステートやContextで管理できるよう整理する。
  - カスタムフックの導入：テーマ切り替えロジックを分離するなど。
  - バックエンドからのWailsイベント（例: ログ受信）を受け取るためのフックや状態管理の枠組み（Zustandなど）を後続ステップで導入しやすい構造にする。

## Capabilities

### New Capabilities
- `core-layout`: アプリの土台となるサイドバー、ヘッダー、詳細ペインなどのUIフレームワークとテーマ管理を提供する基盤部分。
- `log-viewer`: 右側に表示される実行ログ・テレメトリのフィルタリング一覧と、詳細確認領域（DetailPane連携）のUI基盤。
- `dashboard`: デスクトップアプリ起動時のメイン画面。進行中ジョブ・タスクの概要を表示する。

### Modified Capabilities
- なし

## Impact

- `frontend/src/` ディレクトリ内に `components`, `pages` などの標準的なディレクトリ構造が作られます。
- `frontend/src/App.tsx` の表示内容がWailsプレースホルダーから実践的なレイアウトシェルに置き換わります。
- Wails側（Go）との通信はまだモックデータを使用した描画に留まりますが、UIとしては完成状態に近づきます。
