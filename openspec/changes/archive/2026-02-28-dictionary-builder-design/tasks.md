<artifact id="tasks" change="dictionary-builder-design" schema="spec-driven">

## 1. コンポーネントの移植と構成

- [x] 1.1 `mocks-react/src/components/GridEditor.tsx` を `frontend/src/components/dictionary/GridEditor.tsx` にコピー・配置する
- [x] 1.2 `mocks-react/src/pages/DictionaryBuilder.tsx` を `frontend/src/pages/DictionaryBuilder.tsx` にコピー・配置し、インポートパス（`GridEditor` や `DetailPane`、`Layout` など）を適切に修正する
- [x] 1.3 `frontend` プロジェクトの `package.json` にて、`lucide-react` や `@tanstack/react-table` など不足しているパッケージがないか確認し、必要ならインストールする

## 2. ルーティングの統合設定

- [x] 2.1 `frontend/src/App.tsx` （または使用しているルーター定義ファイル）のルーティング設定に `/dictionary` へのルートを追加し、`DictionaryBuilder` コンポーネントを紐付ける
- [x] 2.2 `frontend/src/components/layout/Sidebar.tsx` のナビゲーションアイテム配列に `Dictionary Builder` のリンク (`/dictionary`) と適切なアイコンを追加する

## 3. 動作確認とスタイルの調整

- [x] 3.1 フロントエンド開発サーバーでアプリを表示し、サイドバーから Dictionary Builder 画面へ遷移できるか確認する
- [x] 3.2 辞書ソース一覧（DataTable）での行選択により、DetailPaneが適切にスライドイン・アウト表示されるか視覚的に検証する
- [x] 3.3 エントリー編集(GridEditor)画面に切り替え、セルのインライン編集や行の追加操作が正常に動作するか確認し、コンテナ高さ等にレイアウト崩れがあればTailwindCSSのクラスを修正する

</artifact>

