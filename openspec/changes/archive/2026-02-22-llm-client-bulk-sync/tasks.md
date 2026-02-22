# Tasks: LLM Client Bulk Sync Feature

## 1. 共通インフラストラクチャの開発
- [x] `pkg/infrastructure/llm_client/contract.go` または新規ファイル `bulk.go` に `ExecuteBulkSync` (名前は任意) 関数を追加する。
- [x] 指定された `Concurrency` の数だけ Goroutine ワーカーを起動する処理を実装する。
- [x] 完了したジョブを収集し、入力の順番と対応させて `[]Response` を構築する処理を実装する。
- [x] エラーハンドリング: コンテキストキャンセル時以外の個別エラーは `Response.Success = false` として配列に格納する実装にする。

## 2. コンフィグ連携（UI設定の適用）
- [x] `pkg/config_store` に `llm.bulk_strategy`（値: `"batch"` / `"sync"`）キーを追加し、UI からユーザーがバルク戦略を切り替えられるようにする。
- [x] ローカルLLMプロバイダの場合に `"batch"` の選択を禁止（またはフォールバック）するロジックを実装する。
- [x] `pkg/config_store` に、プラットフォーム毎（Gemini/Local/xAI等）の `sync_concurrency` を取得・保存する設定キーを追加する。
- [x] `LLMManager.GetClient` もしくは呼び出し元で、`llm.bulk_strategy` を読み、`"batch"` なら `GetBatchClient()` / `"sync"` なら `GetClient()` + `ExecuteBulkSync` を選択するルーブリックを実装する。
- [x] デフォルト値（Geminiは5、Localは1等）を定義する。

## 3. ユニットテスト / 結合テスト
- [x] ~~`ExecuteBulkSync` をモッククライアントに対して実行し、正しく設定された並列数以内で実行されるかを検証する。~~ → **スキップ** (`refactoring_strategy.md §6.1` によりユニットテスト排除)
- [x] ~~一部のリクエストがエラーを返した場合でも、全体がクラッシュせずに正しい `Response` リストを構成できるかを検証する。~~ → **スキップ**
- [x] ~~時間のかかるリクエストをキャンセル可能か（コンテキストのキャンセル耐性）を検証する。~~ → **スキップ**
