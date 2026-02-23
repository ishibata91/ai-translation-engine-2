# 本文翻訳 クラス図

## クラス構成

```mermaid
classDiagram
    class BatchTranslator {
        <<interface>>
        +ProposeJobs(ctx context.Context, requests []TranslationRequest, config BatchConfig) (ProposeOutput, error)
        +SaveResults(ctx context.Context, responses []llm_client.Response) error
    }

    class ProposeOutput {
        +[]llm_client.Request Requests
        +[]TranslationResult PreCalculatedResults
    }

    class BatchTranslatorImpl {
        -writer ResultWriter
        -resumeLoader ResumeLoader
        -tagProcessor TagProcessor
        -promptBuilder PromptBuilder
        -bookChunker BookChunker
        -verifier TranslationVerifier
        +ProposeJobs(...)
        +SaveResults(...)
    }

    class ResultWriter {
        <<interface>>
        +Write(result TranslationResult) error
        +Flush() error
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

    class JSONResultWriter {
        -outputDir string
        +Write(result) error
        +Flush() error
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

    class ResumeLoader {
        +LoadCachedResults(config BatchConfig) map~string~TranslationResult
    }

    BatchTranslator <|.. BatchTranslatorImpl : implements
    BatchTranslatorImpl --> ResultWriter : uses
    BatchTranslatorImpl --> ResumeLoader : uses
    BatchTranslatorImpl --> TagProcessor : uses
    BatchTranslatorImpl --> PromptBuilder : uses
    BatchTranslatorImpl --> BookChunker : uses
    BatchTranslatorImpl --> TranslationVerifier : uses
    ResultWriter <|.. JSONResultWriter : implements
    PromptBuilder <|.. DefaultPromptBuilder : implements
    TagProcessor <|.. HTMLTagProcessor : implements
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
        +OutputBaseDir string
        +PluginName string
    }
```

## アーキテクチャの補足：2フェーズモデル (Propose/Save)
本文翻訳（Pass 2）は膨大なレコードを扱うため、JobQueueおよびバッチAPIに最適化した2フェーズモデルを採用している。
- **Phase 1 (Propose)**: `TranslationRequest` を受け取り、差分更新チェック・タグ保護・プロンプト構築を行い、LLMジョブ（リクエスト群）を生成する。既訳やスキップ対象は即時結果として返す。
- **Phase 2 (Save)**: 外部で実行されたLLMのレスポンス群を受け取り、タグ復元・パース・バリデーションを行い、JSONファイルに逐次保存する。

スライス自身は並列通信を管理せず、リクエスト構築と結果の整合性担保（タグ復元等）および永続化に専念する。
