<artifact id="tasks" change="term-translator-slice" schema="spec-driven">

## 1. Core Data Structures & Contract

- [x] 1.1 `pkg/term_translator/contract.go` を作成し、DTO (`TermTranslationRequest`, `TermTranslationResult`, `ReferenceTerm` など) を定義する。
- [x] 1.2 `pkg/term_translator/interfaces.go` または `contract.go` に主要なインターフェース (`TermTranslator`, `TermRequestBuilder`, `TermDictionarySearcher`, `ModTermStore`) を定義する。
- [x] 1.3 `TermRecordConfig` および設定用のドメインモデルを定義する。

## 2. Infrastructure & Utilities

- [x] 2.1 `pkg/term_translator/stemmer.go` を作成し、`github.com/kljensen/snowball` を用いた `KeywordStemmer` を実装する。
- [x] 2.2 `pkg/term_translator/matcher.go` を作成し、探索済みキーワードの重複を排除する `GreedyLongestMatcher` を実装する。
- [x] 2.3 `pkg/term_translator/prompt_builder.go` を作成し、プロンプト生成ロジック (`TermPromptBuilder`) を実装する。

## 3. Database Layer

- [x] 3.1 `pkg/term_translator/store.go` を作成し、Mod用語DBのカプセル化された永続化層 (`SQLiteModTermStore`) を実装する。(テーブル・FTS5仮想テーブル生成、UPSERT 保存処理)
- [x] 3.2 `pkg/term_translator/searcher.go` を作成し、辞書検索ロジック (`SQLiteTermDictionarySearcher`) を実装する。(ExactSearch, KeywordSearch, NPCPartialSearch)

## 4. Translation Logic Execution

- [x] 4.1 `pkg/term_translator/builder.go` を作成し、`ExtractedData` からリクエストを構築する `TermRequestBuilderImpl` を実装する。(NPCペアリング、除外フィルタ連動)
- [x] 4.2 `pkg/term_translator/translator.go` を作成し、メインの翻訳オーケストレーションロジック (`TermTranslatorImpl`) を実装する。(並行処理、検索、プロンプト生成、LLM呼出、結果のストア保存)

## 5. Testing & Validation

- [x] 5.1 `pkg/term_translator/translator_test.go` を作成し、テスト用インメモリDB (`:memory:`) とモックLLMを用いた網羅的パラメタライズドテスト(Table-Driven Tests) を実装する。
- [x] 5.2 全テストケースに OpenTelemetry の TraceID 付きの `context.Context` を渡し、エラーなく実行・完了することを確認する。

## 6. Dependency Injection

- [ ] 6.1 `pkg/term_translator/provider.go` を作成し、Google Wire 用の Provider Set (`TermTranslatorSliceSet`) を定義する。
- [ ] 6.2 プロジェクトルートの `wire.go` を更新して新規プロバイダを登録し、`wire` コマンドを実行して依存関係を生成する。

</artifact>
