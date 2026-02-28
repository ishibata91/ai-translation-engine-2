<artifact id="proposal" change="dictionary-builder-design" schema="spec-driven">

## Why

`mocks-react`で作成された辞書構築ページ（Dictionary Builder）のモックUIを、実際のGo/Wailsアプリケーション（`/frontend`）に移植（ポート）し、実働するコンポーネントとして組み込む必要があります。これにより、ユーザーが実際に辞書データのソースを閲覧・作成・編集し、インライン編集用GridEditorを含むUIで個々の辞書エントリーを管理できるようになります。

## What Changes

- `mocks-react`の `DictionaryBuilder.tsx` や関連コンポーネント群を `/frontend` ディレクトリに移植する
- 共通Layout（`DashboardLayout` や `DetailPane`）を用いた画面構成に統合する
- テーブル表示（TanStack Tableなど）、インライン編集（`GridEditor`）、ソース切り替えなどのUIの結合設計を行う
- UIのルーティングおよびナビゲーションの設定を行う

## Capabilities

### New Capabilities
- `dictionary-builder`: 辞書構築ページのUI機能およびルーティング機能。テーブル描画やGridEditorの表示、DetailPaneを用いた辞書ソース・エントリーの管理UI。

### Modified Capabilities
- `core-layout`: メインナビゲーションからの遷移先に「Dictionary Builder」を追加するなどの調整が必要な場合はルーティング定義を更新。

## Impact

- `/frontend/src/pages/DictionaryBuilder.tsx` などの新規ページコンポーネント
- `/frontend/src/components/dictionary/` コンポーネント群の追加
- `ui_components` などの既存レイアウトやルーティングの変更

</artifact>
