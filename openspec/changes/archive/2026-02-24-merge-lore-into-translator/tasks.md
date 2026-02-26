# タスク一覧: Pass 2 Translator Slice 実装

## 1. 基礎コンポーネントの実装

- [x] `TagProcessor` の実装 (`pkg/translator/tag_processor.go`)
    - [x] 正規表現による HTML タグ抽出とプレースホルダー置換
    - [x] プレースホルダーからオリジナルタグへの復元
    - [x] タグハルシネーションのバリデーションロジック
- [x] `BookChunker` の実装 (`pkg/translator/book_chunker.go`)
    - [x] テキストを指定文字数で分割（HTML 構造維持）
- [x] `ResultWriter` と `ResumeLoader` の実装 (`pkg/translator/persistence.go`)
    - [x] `specs/database_erd.md` に基づく `main_translations` テーブルのマイグレーション実装
    - [x] DB からの既存翻訳結果の読み込み (Resume)
    - [x] 翻訳結果の DB 保存

## 2. Context Engine の統合

- [x] `pkg/translator/context_engine.go` の詳細実装
    - [x] 入力データからの会話コンテキスト抽出
    - [x] 辞書・用語情報の検索ロジック統合
    - [x] 話者プロファイルの取得ロジック
- [x] `Context Engine` を `ProposeJobs` フローに統合

## 3. Phase 1: Propose (翻訳ジョブ生成) の完成

- [x] `pkg/translator/propose.go` の更新
    - [x] `ResumeLoader` によるスキップ処理の追加
    - [x] `TagProcessor` による原文タグ保護の適用
    - [x] `BookChunker` による長文分割の適用
    - [x] `Context Engine` からの情報をプロンプトテンプレートに埋め込む

## 4. Phase 2: Save (翻訳結果保存) の実装

- [x] `pkg/translator/save.go` の新規作成・実装
    - [x] LLM レスポンスからの翻訳文抽出（パース）
    - [x] `TagProcessor` によるタグ復元とバリデーション
    - [x] `ResultWriter` によるDBへの永続化
    - [x] 部分的な失敗時のエラーハンドリング

## 5. 仕上げとテスト

- [x] `pkg/translator/provider.go` の作成（依存性注入の設定）
- [x] `pkg/translator/propose_test.go` の拡充
- [x] `pkg/translator/save_test.go` の新規作成
- [x] Table-Driven Test による各シナリオの網羅（正常系、タグ破損、チャンク結合など）
- [x] `otel` (OpenTelemetry) によるトレースと構造化ログの組み込み
