# Tasks: Persona Gen Job Queue Refactor

## 1. 既存のスライス構造の破棄と再設計
- [ ] `pkg/persona_gen/contract.go` の主力インターフェースを `GeneratePersonas` から、`PreparePrompts(ctx, input Datasets)` と `SaveResults(ctx, input Datasets, results Responses)` の2フェーズ関数へ変更する。
- [ ] `pkg/persona_gen/generator.go` や `provider.go` から、Goroutineによる並列制御、LLMClientのエラーリトライ、プログレス通知（ProgressNotifierへの依存）等の通信ハンドリングコードを完全に削除する。
- [ ] 従来使用していた `LLMManager` への依存を `PersonaGenerator` 構造体および DI用Provider から取り除く。

## 2. Phase 1 実装 (プロンプト生成)
- [ ] 既存の `ContextFilter` (トークン計算等) および `ExcludeAlreadyGenerated` のロジックはそのまま維持し、最終的に `llm_client.Request` の配列だけを構築して返すように `PreparePrompts` を実装する。

## 3. Phase 2 実装 (パースと保存)
- [ ] 渡された `[]llm_client.Response` と元の入力を突き合わせ、正常なレスポンスからペルソナ結果テキストを抽出するロジックを実装する。
- [ ] 抽出された結果を、既存のSQLite層 (`ModTermStore`) を用いて UPSERT するロジックをコールバック関数内に実装する。

## 4. テストスレッドの更新
- [ ] `PersonaGen` 自体のテストからLLMMockを用いた通信・並列処理テストを排除し、「正しいプロンプトが出力されるか」「正しいレスポンス配列が渡された時にDBに正しく保存されるか」の2分割したステートレスなパラメタライズドテストへ書き直す。
