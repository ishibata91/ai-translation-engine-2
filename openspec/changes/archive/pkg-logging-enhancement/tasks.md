# Tasks: pkg 全体へのログガイド第1部適用

## 1. 準備
- [x] [`pkg/infrastructure/telemetry/span.go`](pkg/infrastructure/telemetry/span.go) の `ActionType` 定義に必要なアクション名を追加する。
    - [x] `ActionDBQuery`, `ActionLLMRequest`, `ActionProcessTranslation`, `ActionExport`, `ActionTaskManagement`, `ActionConfigOperation`, `ActionPipelineExecute` 等。

## 2. Infrastructure 層のログ強化
- [x] `pkg/infrastructure/datastore/db.go`: データベースの初期化や設定時に詳細なログを出力するよう更新する。
- [x] `pkg/infrastructure/llm/`: 外部API（Gemini, Local, xAI）呼び出しの前後でリクエスト・レスポンスの内容（機密情報を除く）と所要時間を記録する。
- [x] `pkg/infrastructure/queue/`: ジョブの追加、更新、削除時にステータスと所要時間を記録する。

## 3. ドメイン層（Slices）のログ強化
- [x] `pkg/translator/`: 翻訳処理（Propose, Save）の開始・終了時に詳細なメタデータと所要時間を記録する。
- [x] `pkg/dictionary/`: 辞書エントリの検索、追加、更新時に詳細なログを出力する。
- [x] `pkg/persona/`: ペルソナ生成、スコアリングの各ステップで決定理由を記録する。
- [x] `pkg/terminology/`: 用語抽出、検索の際の詳細情報を記録する。
- [x] `pkg/parser/`: ファイル解析、エンコーディング検出の際の詳細情報を記録する。
- [x] `pkg/summary/`: サマリー生成プロセスの各ステップを記録する。

## 4. 管理・設定層のログ強化
- [x] `pkg/config/`: 設定値の読み書き、マイグレーション実行時に詳細なログを出力する。
- [x] `pkg/task/`: タスクの登録、進捗更新時に詳細なログを出力する。
- [x] `pkg/pipeline/`: パイプライン実行のオーケストレーション（各ステップの遷移）を記録する。

## 5. 仕上げ
- [x] ログメッセージの統一（`span.start`, `span.end`, `error.occurred` 等）を確認する。
- [x] `telemetry.ErrorAttrs(err)...` を漏れなく使用しているか確認する。
- [x] 既存のテストがパスすることを確認し、ログの出力内容を目視で確認する。
