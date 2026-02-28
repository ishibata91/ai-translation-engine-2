## 1. Setup Phase

- [x] 1.1 `react-router-dom` と `zustand` のパッケージを `frontend` にインストールする
- [x] 1.2 `frontend/src/store/uiStore.ts` を作成し、ZustandのUIストア（`theme`, `isSidebarCollapsed`, `logViewerWidth`, `detailPane`状態）を定義する
- [x] 1.3 `frontend/src/hooks/useTheme.ts` を作成し、Zustandのテーマ状態とHTMLの `data-theme` 属性および `localStorage` を同期するフックを実装する

## 2. Component Porting & Refactoring

- [x] 2.1 `frontend/src/components/layout/Header.tsx` を新規作成し、`mocks-react` からヘッダーのUIとテーマ切り替え制御を移植する
- [x] 2.2 `frontend/src/components/layout/Sidebar.tsx` を新規作成し、`mocks-react` からサイドバーを移植、開閉制御を Zustand (`isSidebarCollapsed`) に置き換える
- [x] 2.3 `frontend/src/components/log/LogViewer.tsx` を新規作成し、`mocks-react` から移植、リサイズ時の幅とログ選択時の `DetailPane` トグル制御を Zustand へ置き換える
- [x] 2.4 `frontend/src/components/log/LogDetail.tsx` を `mocks-react` からそのまま移植する
- [x] 2.5 `frontend/src/components/layout/DetailPane.tsx` を新規作成し、シングルトン・ポータルとして動作するように Zustand の状態 (`detailPane.isOpen`, `type`, `payload`) を購読するよう改良する
- [x] 2.6 `frontend/src/components/layout/Layout.tsx` を新規作成し、ルーティングの `Outlet` と上記コンポーネント群 (`Header`, `Sidebar`, `LogViewer`, `DetailPane`) を適切に配置するコンテナを構築する
- [x] 2.7 `frontend/src/pages/Dashboard.tsx` を `mocks-react` から移植する

## 3. App Integration & Validation

- [x] 3.1 `frontend/src/App.tsx` を改修し、`react-router-dom`の `HashRouter` で `Layout.tsx` と `Dashboard.tsx` をルーティングするよう設定する
- [x] 3.2 アプリをビルド・起動し、UIテーマ切り替え、サイドバーの開閉、ルーティング、LogViewerのリサイズ機能がWails環境上で想定通り動作するか確認する
- [x] 3.3 LogViewer から適当なログをクリックし、DetailPane が下部からスライドアップして詳細が表示されるか単体テストする

## 4. Documentation Update (If necessary)

- [x] 4.1 UIの主要な構造変更について、必要に応じて `specs/architecture.md` 等の共通ドキュメントに補足説明を追記する
