<artifact id="design" change="dictionary-builder-design" schema="spec-driven">

## Context

現在、`mocks-react` プロジェクトにおいて `DictionaryBuilder.tsx` や `GridEditor.tsx` のモックインターフェースが実装されています。この画面は、辞書データのソース一覧表示や追加機能、および個別のソース内容（辞書エントリー群）のインライン編集（ExcelライクなグリッドUI）を行うための画面です。これを Wails バックエンドと協調して動作する本番環境 (`/frontend` プロジェクト) に移植し、他の画面（Dashboard等）と統一されたレイアウト (Layout や DetailPane) 内で統合する必要があります。

## Goals / Non-Goals

**Goals:**
- `mocks-react` における Dictionary Builder のUI機能（辞書ソース管理、エントリーのインライン編集用の `GridEditor` など）を `/frontend` に移植する。
- 既存の共通コンポーネントである `Layout.tsx`、`Sidebar.tsx` と統合し、ルーティングによって正しく画面遷移が行われるようにする。
- 個別ソースの編集モードや、詳細を表示する `DetailPane` との連携を確保し、UXを維持する。

**Non-Goals:**
- Wails バックエンドとのデータの実際の連携 (Go / `Bindings`) は本段階のUI移植設計以降の別設計・タスクとし、本フェーズではUIレイアウトとモックデータの統合に留める。
- 本件とは無関係なページ（Master Persona や Translation Flow など）の移植は対象外。

## Decisions

- **移植先ディレクトリ構成**: 
  - `mocks-react/src/pages/DictionaryBuilder.tsx` -> `/frontend/src/pages/DictionaryBuilder.tsx`
  - `mocks-react/src/components/GridEditor.tsx` -> `/frontend/src/components/dictionary/GridEditor.tsx`
  これにより、辞書構築機能に関するコンポーネントをわかりやすく整理します。
- **UI統合・ルーティング**: `App.tsx` の `react-router-dom` を更新し、`/dictionary` エンドポイントを定義。さらに `Sidebar.tsx` などのナビゲーションから同ルートへ遷移可能にします。
- **DetailPaneとレイアウト設計**: `Layout` コンポーネントを共通のラッパーとして用い、テーブルで辞書ソースを選択した際の詳細情報表示は、既存の `/frontend/src/components/layout/DetailPane.tsx` などを活用して行います。

## Risks / Trade-offs

- **[Risk]** `mocks-react` 環境から移植する際、TailwindCSS の設定や全体コンテナの幅・高さ設定の差異により、グリッドが画面内に収まらないなどのレイアウト崩れが発生する可能性がある。
  -> **Mitigation**: 移植後、ブラウザで実際に表示し、flex 構造と height (`h-full` など) が正しく継承されているか目視確認し、必要なスタイル調整をページレベルで行う。
- **[Risk]** 一部のアイコンや UI コンポーネントに不足が生じる場合がある。
  -> **Mitigation**: `lucide-react` などの必要な UI ライブラリが `/frontend/package.json` に含まれているか確認し、不足分は速やかに追加する。

</artifact>
