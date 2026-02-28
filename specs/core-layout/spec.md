# Core Layout

## Purpose
TBD: フロントエンドの全体レイアウト、テーマ管理、サイドバーナビゲーションに関する振る舞いを定義します。

## Requirements

### Requirement: Layout Base Structure
フロントエンドアプリのUIのベースとして、ヘッダー、サイドバー、メインコンテンツ領域を備えたレイアウトコンポーネント構造を提供しなければならない。

#### Scenario: App Initialization
- **WHEN** アプリケーションが起動される
- **THEN** ヘッダー、サイドバー、およびコンテンツ領域が正しく描画される

### Requirement: Theme Selection
ユーザーはUIのカラーテーマ（dark, lightなど）を選択・切り替えできなければならない。テーマは `localStorage` に保存され再起動時に復元される。

#### Scenario: Change Theme
- **WHEN** テーマ切り替えのドロップダウンから新しいテーマ（例: cupcake）を選択する
- **THEN** アプリ全体のHTMLルート要素の `data-theme` 属性が更新される
- **AND** `localStorage` の値が更新される

### Requirement: Collapsible Sidebar Navigation
サイドバーは開閉可能であり、ルーティングに応じたアクティブ状態を表示できなければならない。

#### Scenario: Toggle Sidebar
- **WHEN** サイドバーの開閉トグルボタンをクリックする
- **THEN** サイドバーの幅が縮小され、テキストラベルが非表示になりアイコンのみになる
- **AND** サイドバーの幅に応じてメインコンテンツ領域が自動的にリサイズされる
