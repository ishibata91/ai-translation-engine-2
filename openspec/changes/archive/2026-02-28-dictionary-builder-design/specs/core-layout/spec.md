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

### Requirement: Sidebar Navigation Items Update
サイドバーのナビゲーションアイテムに「Dictionary Builder」へのリンクが追加され、クリック可能でなければならない（SHALL）。

#### Scenario: Navigate to Dictionary Builder
- **WHEN** ユーザーがサイドバーの「Dictionary Builder」ナビゲーション項目をクリックしたとき
- **THEN** ルーターが `/dictionary` に遷移し、それに伴うページが描画されること
- **AND** サイドバーのその項目がアクティブ状態として表示されること

### Requirement: ダークモード対応の確保
アプリケーション全体の要素（画面内のアラート表示や、コードタグ等のインライン要素など）において、ダークモード時でも十分なコントラスト比が確保され、テキストが可読でなければならない（SHALL）。
