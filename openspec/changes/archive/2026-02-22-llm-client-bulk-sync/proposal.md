# Proposal: LLM Client Bulk Sync Feature

## Problem & Motivation
現在、LLM Client は単一の `Request` に対する同期的な `Complete` メソッドと、非同期な `SubmitBatch` メソッドを提供しています。
しかし、スライス（例：SummaryGeneratorSlice）側で複数のプロンプトを生成し、それを同期的に一括処理したい場合、スライス側でGoroutineを用いた独自の並列ループ制御やエラーハンドリングを実装せざるを得ません。
これはVSAの原則（ドメインロジックへの集中）に反し、インフラ固有のロジックがスライスに漏れ出しています。また、UIコンフィグ（設定画面）から「APIの並列実行数」を動的に制御する仕組みも欠如しています。

## Proposal
インフラストラクチャ層の `LLMClient` インターフェースを拡張し、複数のリクエストを一括で受け取り、指定された並列数で処理を完了させて結果を返す `CompleteBulk`（または同等の）メソッドを追加します。
この並列数は、UIからのコンフィグ設定（ConfigStore経由）で動的に受け取れるようにし、APIレート制限への安全性と処理効率のバランスをユーザーが制御できるようにします。

## Capabilities

### New Capabilities
- `LLMClientBulk`: LLM Clientでの同期的なバルクリクエスト並列処理機能。

### Modified Capabilities
- `LLMClient`: `CompleteBulk` インターフェースの追加と並列数制御の組み込み。

## Impact
- `pkg/infrastructure/llm_client/contract.go` のインターフェース拡張
- 各LLMプロバイダ（Gemini, OpenAI, xAI, Local）のクライアント実装における並列ワーカープールの追加
- `ConfigStore` への「LLM並列実行数」設定値の追加
