## Why

xAI Batch 実行で返却順が入力順と一致しない場合に、レスポンスが別リクエストへ誤って保存される不具合がある。Gemini では仕様上入力順が返るが、Queue 側が index 前提のままだと将来経路や再開順差異に弱いため、provider 共通で相関 ID ベースへ統一する必要がある。

## What Changes

- xAI Batch 結果を配列順で扱う前提を廃止し、`batch_request_id` によるリクエスト相関を必須化する。
- Gemini Batch でも `inlinedRequest.metadata` / `inlinedResponse.metadata` に同一相関 ID を通し、LLM本文に依存しない対応付けを明示する。
- Queue Worker の batch 結果保存を index 対応から相関 ID 対応へ変更し、順不同レスポンスでも正しい job に保存する。
- batch 結果欠落・重複・未知 ID を検知したときの失敗扱いを定義し、誤保存より安全側で終了する。
- Resume 時の処理順差異（DB 取得順や provider 返却順）に影響されないことを、テスト要件として追加する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `llm`: Provider-native Batch API の結果取得で、入力順ではなく request 相関キー（`queue_job_id`）に基づいて結果を復元しなければならない。xAI は `batch_request_id`、Gemini は `metadata` 経路で同一キーを伝播しなければならない。
- `slice/queue`: Queue の batch 結果適用は `jobs[i] <- results[i]` 前提を持たず、相関 ID ベースで request 状態を更新しなければならない。

## Impact

- 影響コード:
  - `pkg/gateway/llm/xai_client.go`
  - `pkg/gateway/llm/gemini_batch_client.go`
  - `pkg/runtime/queue/worker.go`
  - `pkg/runtime/queue/queue.go`（必要に応じた取得順安定化）
  - `pkg/gateway/llm/xai_batch_client_test.go`
  - `pkg/gateway/llm/gemini_batch_client_test.go`
  - `pkg/runtime/queue/job_queue_test.go`
- API/データ:
  - 外部 API 仕様変更はなし。
  - DB スキーマ変更は原則なし（必要時のみ `database_erd.md` へ影響を記載）。
- システム影響:
  - Master Persona の batch 実行・再開時の結果整合性が向上する。
  - 誤保存リスクを低減する代わりに、不整合検知時は明示的に失敗として扱う。