# Log Viewer

## Purpose
TBD: システムログやテレメトリを表示するログビューアと、関連する詳細パネルの振る舞いを定義します。

## Requirements

### Requirement: Log Viewer Visibility and Resizing
システム状態監視のために右側に配置される「ログビューア」は、手動で横幅をリサイズ可能でなければならない。

#### Scenario: Resize Log Viewer
- **WHEN** ログビューアの左端にあるリサイズ用ハンドルをドラッグする
- **THEN** ログビューアの横幅がマウス移動に合わせて変化する（最小200px、最大800px）

### Requirement: Log Filtering
表示中のログリストは、ログレベル（INFO, WARN, ERROR等）とTraceIDによる文字列で絞り込みができなければならない。

#### Scenario: Filter by Error Level
- **WHEN** ログレベルの選択UIで「ERROR」を選択する
- **THEN** リスト内にエラーログの項目のみが表示される

### Requirement: Log Detail Pane Integration
ログのエントリを選択した際、画面下部から詳細情報を表示するパネル（Detail Pane）を展開できなければならない。

#### Scenario: View Log Details
- **WHEN** ログビューア内の任意のログエントリをクリックする
- **THEN** 画面下部よりDetailPaneがスライドアップ・アニメーションを伴って展開される
- **AND** パネル内にそのログのエラーメッセージやスタックトレース詳細等が表示され、タイトルがクリックしたログのものに更新される
