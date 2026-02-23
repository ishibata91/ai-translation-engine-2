# Tasks: Term Translator Job Queue Refactor

## 1. 既存のスライス構造の破棄と再設計
- [x] **[破壊的変更]** 既存の `GenerateTranslations` を含む、LLMとの同期通信や待機に関連するすべての関数を破棄する。唯一、プロンプト構築ロジック（`buildPrompt` 等）のみを Phase 1 用に抽出・再利用の対象とする。
- [x] `pkg/term_translator/contract.go` 内の主力インターフェースを、2フェーズ関数 `PreparePrompts(ctx, input Datasets)` と `SaveResults(ctx, input Datasets, results Responses)` へ全面的に書き換える。
- [x] `pkg/term_translator/translator.go` などから、Goroutine による完了待機、チャネル制御、エラーリトライ、および `LLMManager` への直接依存を完全に削除する。

## 2. Phase 1 実装 (プロンプト生成)
- [x] 既存の `buildPrompt` 等のロジックを流用し、入力された `TermTranslatorInput` の各要素に対して辞書引きとコンテキスト構築を行い、`llm_client.Request` の配列を純粋に返すように実装を修正する。

## 3. Phase 2 実装 (パースと保存)
- [x] 渡された `[]llm_client.Response` と元の入力を突き合わせ、正常なレスポンスから翻訳結果テキスト（`TL: |xxx|`）を抽出するロジックを実装する。
- [x] 抽出された結果を、既存のSQLite層 (`ModTermStore`) を用いて UPSERT するロジックをコールバック関数内に実装する。

## 4. テストスレッドの更新
- [x] `TermTranslator` 自体のテストからLLMMockを用いた通信テストを排除し、「正しいプロンプトが出力されるか」「正しいレスポンス配列が渡された時にDBに正しく保存されるか」の2分割したステートレスなパラメタライズドテストへ書き直す。

## 5. 構造化ログと TraceID の実装
- [x] `PreparePrompts` および `SaveResults` の各フェーズで、`specs/refactoring_strategy.md` に基づく Entry/Exit ログを `slog` を用いて実装する。
- [x] 全てのログ出力において、`context.Context` を介して TraceID / SpanID が伝播されるようにする。
- [x] パース失敗やエラーレスポンスの際、詳細なコンテキストを含めた `debug` または `warn` ログを出力するように実装する。
