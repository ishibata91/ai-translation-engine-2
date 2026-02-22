# Design: LLM Client Bulk Sync Feature

## Context
垂直スライスアーキテクチャ(VSA)において、スライスはドメインロジックへの集中が求められます。しかし現在、各ドメインスライスが複数レコードを処理する際、LLMClientに対して「N件のプロンプトを投げて、すべて完了するまで待機する（並列処理・チャネル制御等のインフラレイヤーのコードを含む）」というインフラ制御の責務をスライス内部に記述する形になってしまっています。
さらに、同期バッチ処理時のレートリミット回避やパフォーマンス最適化のための「API並列実行数」を、ユーザーがUI（ConfigStore）から動的に設定・適用する仕組みが存在しません。

また、クラウド系プロバイダ（Gemini, xAI 等）はプロバイダ提供の非同期 Batch API を活用できるが、ローカルLLM（GGUF/Local）はバッチAPIを持たないため同期完了（`ExecuteBulkSync`）にフォールバックする必要がある。どちらを使うかをユーザーが UI から切り替えられる仕組みが必要である。

## Goals / Non-Goals

**Goals:**
- LLMClientパッケージ内に同期的なバルク処理機能（`ExecuteBulkSync` ヘルパー）を追加し、複数リクエストのGoroutine並列実行と結果集約をインフラ層へ隠蔽する。
- バルク処理戦略（`"batch"` = プロバイダ提供の非同期 Batch API / `"sync"` = `ExecuteBulkSync` による同期完了）を ConfigStore 経由でユーザーが選択できるようにする。ローカルLLMは `"batch"` 非対応のため `"sync"` のみ有効。
- `LLMConfig` または専用の引数として「並列実行数（Concurrency）」を受け取り、ワーカープール等でAPIへの流量制御を確実に行う（`"sync"` 戦略時に有効）。

**Non-Goals:**
- OpenAIやxAIの提供する「Batch API（非同期で数分〜24時間かかる仕組み）」のポーリングや状態管理機能の実装。これは後続の `JobQueue Infrastructure` の責務として別Changeで対応する。
- 再試行（Retry）ロジック自体の変更。既存の `retry.go` は引き続き個別の `Complete` 単位で有効に機能するものとする。
- 各ドメインスライス（や今後作成するJobQueue等）から、チャネルやGoroutineを使った複雑な並列リクエスト管理コードを排除する作業。これは `ExecuteBulkSync` の整備が完了した後、別Changeで段階的に対応する。

## Decisions

1. **インフラ層での汎用ワーカープール実装**:
   - 各プロバイダの `LLMClient` 具象クラスに個別実装するのはDRY原則違反となるため、`llm_client` パッケージの共通ヘルパー関数 `ExecuteBulkSync(ctx, client, reqs, concurrency)` としてワーカープールパターンを実装する。

2. **バルク戦略の ConfigStore 制御**:
   - ConfigStore に `llm.bulk_strategy` キー（値: `"batch"` / `"sync"`）を追加し、UI からユーザーが切り替えられるようにする。
   - `LLMManager`（または呼び出し元の ProcessManager / JobQueue）がこのキーを読み、`"batch"` なら `GetBatchClient()` → `SubmitBatch`、`"sync"` なら `GetClient()` → `ExecuteBulkSync` を選択する。
   - ローカルLLMプロバイダが選択されている場合は `"batch"` を選んでも強制的に `"sync"` へフォールバックし、UI 上でもその旨を表示する（設定値の上書きは行わない）。

3. **UIからの並列数設定（`"sync"` 戦略時）**:
   - `LLMConfig` に `Concurrency int` フィールドを追加。
   - ConfigStore のキー `llm.sync_concurrency` から読み出し、未設定時はプロバイダごとのデフォルト値（ローカル: 1、Gemini: 5 等）にフォールバックする。`"batch"` 戦略時はこの値は使用しない。

4. **個別エラーの許容（Partial Failure）**:
   - 100件中1件がエラー（レートリミット超過の永続エラーやセーフティブロック等）になっても、残り99件の成功結果は失われない仕様とする。
   - `ExecuteBulkSync` 自身はクリティカルな例外（`ctx.Done()` など）以外では `error` を返さず、返り値の `[]Response` の各要素の `Success` フラグと `Error` メッセージで個別の成否を表現する。

## Risks / Trade-offs

- **[Risk] 大量リクエスト時のメモリ消費**:
  数万件の `Request` 構造体を一度にメモリに乗せると、メモリ領域を圧迫する可能性がある。
  → **Mitigation**: 呼び出し元（ProcessManagerやJob Queue等）の層で、常識的なチャンクサイズ（例：500件〜1000件）ごとに分割して `ExecuteBulkSync` を呼び出す運用とし、インフラ層のこのヘルパーは渡されたスライスを純粋に処理することに徹する。
