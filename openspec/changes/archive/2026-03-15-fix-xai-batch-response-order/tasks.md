## 1. 相関ID規約の導入

- [x] 1.1 `runtime/queue` で batch submit 前に各 request metadata へ `queue_job_id=<job.ID>` を注入する
- [x] 1.2 `queue_request_seq` の扱い（診断用途）を実装方針どおり整理する
- [x] 1.3 相関IDが設定される経路（新規実行/再開）をログで追跡できるようにする

## 2. xAI Batch 相関対応

- [x] 2.1 `xai_client.go` の `addRequests` で `metadata.queue_job_id` を `batch_request_id` に優先採用する
- [x] 2.2 `xai_client.go` の `parseResults` で `batch_request_id` を `Response.Metadata.queue_job_id` へ復元する
- [x] 2.3 `batch_request_id` 欠落時のエラー/フォールバックを定義どおり実装する

## 3. Gemini Batch 相関対応

- [x] 3.1 `gemini_batch_client.go` の submit で `inlinedRequest.metadata.queue_job_id` が送信されることを保証する
- [x] 3.2 `gemini_batch_client.go` の result parse で `Response.Metadata.queue_job_id` を維持する
- [x] 3.3 Gemini 経路で LLM本文に依存しない相関であることをコードコメント/ログで明確化する

## 4. Queue 結果適用ロジックの更新

- [x] 4.1 `applyBatchResults` を metadata-first（`queue_job_id` 優先）へ変更する
- [x] 4.2 未知ID・重複ID・欠落IDの失敗処理を追加し、誤保存を防止する
- [x] 4.3 metadata 欠落時の互換 fallback（index 適用）を限定的に維持する

## 5. テスト追加

- [x] 5.1 `xai_batch_client_test.go` に順不同結果でも相関復元できるケースを追加する
- [x] 5.2 `gemini_batch_client_test.go` に metadata round-trip の相関検証ケースを追加する
- [x] 5.3 `job_queue_test.go` に結果シャッフル/再開順差異でも正しい job 更新になるケースを追加する

## 6. 品質ゲート

- [x] 6.1 変更ファイルごとに `npm run backend:lint:file -- <file...>` を実行する
- [x] 6.2 修正後に対象ファイルで `npm run backend:lint:file -- <file...>` を再実行する
- [x] 6.3 `go test ./pkg/gateway/llm ./pkg/runtime/queue` を実行する
- [x] 6.4 最終確認として `npm run lint:backend` を実行する
