# Tasks: Term Translator Job Queue Refactor

## 1. 既存のスライス構造の破棄と再設計
- [ ] `pkg/term_translator/contract.go` 内の主力インターフェースを、一貫して実行する `GenerateTranslations` から、`PreparePrompts(ctx, input Datasets)` と `SaveResults(ctx, input Datasets, results Responses)` の2フェーズ関数へ変更する。
- [ ] `pkg/term_translator/translator.go` などに散らばるGoroutineによるLLM完了待機ロジックや、LLMClientのエラーリトライ等の通信ハンドリングコードを完全に削除する。
- [ ] 従来使用していた `LLMManager` への依存を `TermTranslator` 構造体および DI用Provider から取り除く。

## 2. Phase 1 実装 (プロンプト生成)
- [ ] 既存の `buildPrompt` 等のロジックを流用し、入力された `TermTranslatorInput` の各要素に対して辞書引きとコンテキスト構築を行い、`llm_client.Request` の配列を純粋に返すように実装を修正する。

## 3. Phase 2 実装 (パースと保存)
- [ ] 渡された `[]llm_client.Response` と元の入力を突き合わせ、正常なレスポンスから翻訳結果テキスト（`TL: |xxx|`）を抽出するロジックを実装する。
- [ ] 抽出された結果を、既存のSQLite層 (`ModTermStore`) を用いて UPSERT するロジックをコールバック関数内に実装する。

## 4. テストスレッドの更新
- [ ] `TermTranslator` 自体のテストからLLMMockを用いた通信テストを排除し、「正しいプロンプトが出力されるか」「正しいレスポンス配列が渡された時にDBに正しく保存されるか」の2分割したステートレスなパラメタライズドテストへ書き直す。
