## MODIFIED Requirements

### Requirement: Log Filtering
表示中のログリストは、ログレベル（INFO, WARN, ERROR等）とTraceIDによる文字列で絞り込みができなければならない。また、システム初期化時のデフォルトのログレベルは「ERROR」とし、ユーザーが選択したログレベルは再起動後も保持されなければならない。

#### Scenario: Filter by Error Level
- **WHEN** ログレベルの選択UIで「ERROR」を選択する
- **THEN** リスト内にエラーログの項目のみが表示される

#### Scenario: Default Log Level is Error
- **WHEN** アプリケーションを初回起動する
- **THEN** ログレベルフィルターの初期値が「ERROR」に設定されている

#### Scenario: Persist Selected Log Level
- **WHEN** ログレベルを「WARN」に変更し、アプリケーションを再起動する
- **THEN** 再起動後もログレベルフィルターが「WARN」に維持されている

### Requirement: Log Detail Pane Integration
ログのエントリを選択した際、画面下部から詳細情報を表示するパネル（Detail Pane）を展開できなければならない。詳細情報にはバックエンドのテレメトリ属性（ActionType, ResourceType, trace_id, その他カスタム属性）が同期して表示されなければならない。

#### Scenario: View Log Details
- **WHEN** ログビューア内の任意のログエントリをクリックする
- **THEN** 画面下部よりDetailPaneがスライドアップ・アニメーションを伴って展開される
- **AND** パネル内にそのログのメッセージ、スタックトレース、および各種テレメトリ属性（trace_id等）が表示され、タイトルがクリックしたログのものに更新される

## ADDED Requirements

### Requirement: Asynchronous Log Streaming
バックエンド（Go）で発生したテレメトリログは、実処理をブロックすることなく非同期でフロントエンド（UI）のログビューワーにストリーミングされなければならない。

#### Scenario: Receive Log Asynchronously
- **WHEN** バックエンドで新しいログエントリが記録される
- **THEN** UIのログビューワー内に新しいログが非同期で追加表示される
- **AND** バックエンドの処理パフォーマンスに遅延を発生させない

### Requirement: TraceID Consistency
バックエンドのテレメトリ実装において、リクエストを一意に特定するための識別子の名称は「trace_id」に統一されなければならない（従来の「request_id」は「trace_id」に変更される）。

#### Scenario: Verify TraceID Name
- **WHEN** バックエンドからUIにログがストリーミングされる
- **THEN** ログの構造化データにおいて、リクエスト識別子が `trace_id` というキーで格納されている
