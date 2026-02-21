# 本文翻訳 クラス図

## クラス構成

```mermaid
classDiagram
    class BatchTranslator {
        <<interface>>
        +TranslateBatch(ctx context.Context, requests []TranslationRequest, config BatchConfig) ([]TranslationResult, error)
    }

    class BatchTranslatorImpl {
        -translator Translator
        -writer ResultWriter
        -resumeLoader ResumeLoader
        +TranslateBatch(ctx, requests, config)
        -loadCachedResults(config BatchConfig) map~string~TranslationResult
        -buildRequestKey(req TranslationRequest) string
    }

    class Translator {
        <<interface>>
        +Translate(ctx context.Context, request TranslationRequest) (TranslationResult, error)
    }

    class ResultWriter {
        <<interface>>
        +Write(result TranslationResult) error
        +Flush() error
    }

    class TranslatorImpl {
        -promptBuilder PromptBuilder
        -tagProcessor TagProcessor
        -llmClient LLMClient
        -verifier TranslationVerifier
        -retryConfig RetryConfig
        +Translate(ctx, request) (TranslationResult, error)
        -translateWithRetry(ctx, req) (string, error)
        -translateBook(ctx, req) (string, error)
    }

    class JSONResultWriter {
        -outputDir string
        -mu sync.Mutex
        -buffers map~string~[]TranslationResult
        +Write(result) error
        +Flush() error
        -writeToFile(plugin string) error
    }

    class PromptBuilder {
        <<interface>>
        +Build(request TranslationRequest) (systemPrompt string, userPrompt string, err error)
    }

    class TagProcessor {
        <<interface>>
        +Preprocess(text string) (processedText string, tagMap map~string~string)
        +Postprocess(text string, tagMap map~string~string) string
        +Validate(translatedText string, tagMap map~string~string) error
    }

    class TranslationVerifier {
        <<interface>>
        +Verify(sourceText string, translatedText string, tagMap map~string~string) error
    }

    class BookChunker {
        <<interface>>
        +Chunk(text string, maxTokensPerChunk int) []string
    }

    class DefaultPromptBuilder {
        -configStore ConfigStore
        +Build(request) (string, string, error)
    }

    class HTMLTagProcessor {
        +Preprocess(text) (string, map~string~string)
        +Postprocess(text, tagMap) string
        +Validate(translatedText, tagMap) error
    }

    class DefaultTranslationVerifier {
        +Verify(source, translated, tagMap) error
    }

    class HTMLBookChunker {
        -tokenCounter func(string) int
        +Chunk(text, maxTokens) []string
        -splitByParagraph(text) []string
        -splitBySentence(text) []string
    }

    BatchTranslator <|.. BatchTranslatorImpl : implements
    BatchTranslatorImpl --> Translator : uses
    BatchTranslatorImpl --> ResultWriter : uses
    Translator <|.. TranslatorImpl : implements
    ResultWriter <|.. JSONResultWriter : implements
    TranslatorImpl --> PromptBuilder : uses
    TranslatorImpl --> TagProcessor : uses
    TranslatorImpl --> LLMClient : uses
    TranslatorImpl --> TranslationVerifier : uses
    TranslatorImpl --> BookChunker : uses
    PromptBuilder <|.. DefaultPromptBuilder : implements
    TagProcessor <|.. HTMLTagProcessor : implements
    TranslationVerifier <|.. DefaultTranslationVerifier : implements
    BookChunker <|.. HTMLBookChunker : implements
```

## DTO定義

```mermaid
classDiagram
    class TranslationResult {
        +ID string
        +RecordType string
        +SourceText string
        +TranslatedText *string
        +Index *int
        +Status string
        +ErrorMessage *string
        +SourcePlugin string
        +SourceFile string
        +EditorID *string
    }

    class BatchConfig {
        +MaxWorkers int
        +TimeoutSeconds float64
        +MaxTokens int
        +OutputBaseDir string
        +PluginName string
    }

    class RetryConfig {
        +MaxRetries int
        +BaseDelaySeconds float64
        +MaxDelaySeconds float64
        +ExponentialBase float64
    }

    class TagHallucinationError {
        +Message string
        +Error() string
    }
```

## 依存関係

- `BatchTranslatorImpl` → `Translator`: 単一リクエストの翻訳実行
- `BatchTranslatorImpl` → `ResultWriter`: 翻訳結果の逐次保存
- `BatchTranslatorImpl` → `ResumeLoader`: 既存翻訳結果の読み込み（差分更新）
- `TranslatorImpl` → `PromptBuilder`: プロンプト構築
- `TranslatorImpl` → `TagProcessor`: HTMLタグ前処理/後処理
- `TranslatorImpl` → `LLMClient` (共通インフラ): LLM呼び出し
- `TranslatorImpl` → `TranslationVerifier`: 翻訳結果の品質検証
- `TranslatorImpl` → `BookChunker`: 書籍長文分割
- `DefaultPromptBuilder` → Config Store: プロンプトテンプレートの取得
- `JSONResultWriter` → ファイルシステム: JSON出力
- Process Manager → `BatchTranslator`: バッチ翻訳の起動
