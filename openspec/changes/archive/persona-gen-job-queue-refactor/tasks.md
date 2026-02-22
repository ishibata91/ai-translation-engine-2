# Tasks: Persona Gen Job Queue Refactor

## 1. 既存のスライス構造の破棄と再設計
- [x] `pkg/persona_gen/contract.go` の主力インターフェースを `GeneratePersonas` から、`PreparePrompts(ctx context.Context, input Datasets)` と `SaveResults(ctx context.Context, input Datasets, results Responses)` の2フェーズ関数へ変更する。
- [x] `pkg/persona_gen/generator.go` や `provider.go` から、Goroutineによる並列制御、LLMClientのエラーリトライ、プログレス通知（ProgressNotifierへの依存）等の通信ハンドリングコードを完全に削除する。
- [x] 従来使用していた `LLMManager` への依存を `PersonaGenerator` 構造体および DI用Provider から取り除く。

## 2. ログとテストの基盤整備 (Refactoring Strategy 準拠)
- [x] すべて de Contract メソッドの入り口と出口で `slog.DebugContext` を用いた引数・戻り値の記録（Entry/Exitログ）を実装する。
- [x] `pkg/persona_gen/generator_test.go` を更新し、テスト専用の SQLite (:memory:) を使用した網羅的なパラメタライズドテスト（Table-Driven Test）を構築する。細粒度なユニットテストは排除する。

## 3. Phase 1 実装 (プロンプト生成)
- [x] 既存の `ContextFilter` (トークン計算等) および `ExcludeAlreadyGenerated` のロジックはそのまま維持し、最終的に `llm_client.Request` の配列だけを構築して返すように `PreparePrompts` を実装する。

## 4. Phase 2 実装 (パースと保存)
- [x] 渡された `[]llm_client.Response` と元の入力を突き合わせ、`TL: |...|` 形式からペルソナ結果テキストを抽出するロジックを実装する。
    - [x] 4.1. `Metadata` から NPC ID を取得し、各レスポンスを識別するロジックの実装
    - [x] 4.2. 正規表現およびフォールバックロジックを用いたペルソナ抽出機能の実装
    - [x] 4.3. 抽出結果が空または不正な場合のバリデーション実装
- [x] 抽出された結果を、既存のSQLite層 (`ModTermStore`) を用いて UPSERT するロジックをコールバック関数内に実装する。
    - [x] 4.4. `term_key` の命名規則 (`NPC_PERSONA_<NPC_ID>`) に基づくキー生成の実装
    - [x] 4.5. `ModTermStore.Upsert` を用いた永続化処理の実装
    - [x] 4.6. 個別の保存失敗を許容し、継続するループ構造の実装
    - [x] 4.7. 成功・失敗数に応じた `slog` による構造化ログ出力の実装
