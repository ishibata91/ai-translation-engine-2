# 用語翻訳・Mod用語DB保存 クラス図

```mermaid
classDiagram
    class TermTranslatorInput {
        +[]NPC NPCs
        +[]Item Items
        +[]Magic Magics
        +[]Message Messages
        +[]Location Locations
        +[]Quest Quests
    }

    class ProposeOutput {
        +[]llm_client.Request Requests
        +[]TermTranslationResult PreCalculatedResults
    }

    class TermTranslationRequest {
        +String FormID
        +String EditorID
        +String RecordType
        +String SourceText
        +String ShortName
        +String SourcePlugin
        +String SourceFile
        +[]ReferenceTerm ReferenceTerms
    }

    class ReferenceTerm {
        +String Source
        +String Translation
    }

    class TermTranslationResult {
        +String FormID
        +String EditorID
        +String RecordType
        +String SourceText
        +String TranslatedText
        +String SourcePlugin
        +String SourceFile
        +String Status
        +String ErrorMessage
    }

    class TermTranslator {
        <<Interface>>
        +ProposeJobs(ctx context.Context, input TermTranslatorInput) (ProposeOutput, error)
        +SaveResults(ctx context.Context, responses []llm_client.Response) error
    }

    class TermRequestBuilder {
        <<Interface>>
        +BuildRequests(ctx context.Context, input TermTranslatorInput) ([]TermTranslationRequest, error)
    }

    class TermDictionarySearcher {
        <<Interface>>
        +SearchExact(ctx context.Context, text string) ([]ReferenceTerm, error)
        +SearchKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error)
        +SearchNPCPartial(ctx context.Context, keywords []string, consumedKeywords []string, isNPC bool) ([]ReferenceTerm, error)
        +SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error)
    }

    class ModTermStore {
        <<Interface>>
        +InitSchema(ctx context.Context) error
        +SaveTerms(ctx context.Context, results []TermTranslationResult) error
        +GetTerm(ctx context.Context, originalEN string) (string, error)
    }

    class TermTranslatorImpl {
        -TermRequestBuilder requestBuilder
        -TermDictionarySearcher dictSearcher
        -ModTermStore modTermStore
        -TermPromptBuilder promptBuilder
        +ProposeJobs(ctx context.Context, input TermTranslatorInput) (ProposeOutput, error)
        +SaveResults(ctx context.Context, responses []llm_client.Response) error
    }

    class TermRequestBuilderImpl {
        -TermRecordConfig config
        +BuildRequests(ctx context.Context, input TermTranslatorInput) ([]TermTranslationRequest, error)
    }

    class TermRecordConfig {
        +[]string TargetRecordTypes
        +IsTarget(recordType string) bool
    }

    class SQLiteTermDictionarySearcher {
        -db *sql.DB
        -[]string additionalDBPaths
        +SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error)
    }

    class KeywordStemmer {
        +Stem(word string) string
        +StripPossessive(word string) string
    }

    class SQLiteModTermStore {
        -db *sql.DB
        +InitSchema(ctx context.Context) error
        +SaveTerms(ctx context.Context, results []TermTranslationResult) error
    }

    class TermPromptBuilder {
        +BuildSystemPrompt(recordType string) string
        +BuildUserPrompt(request TermTranslationRequest) string
    }

    class ProcessManager {
        -TermTranslator termTranslator
        -JobQueue jobQueue
        +HandleTermTranslation(w http.ResponseWriter, r *http.Request)
    }

    ProcessManager --> TermTranslator : uses
    TermTranslator <|.. TermTranslatorImpl : implements
    TermTranslatorImpl --> TermRequestBuilder : uses
    TermTranslatorImpl --> TermDictionarySearcher : uses
    TermTranslatorImpl --> ModTermStore : uses
    TermTranslatorImpl --> TermPromptBuilder : uses
    TermRequestBuilder <|.. TermRequestBuilderImpl : implements
    TermRequestBuilderImpl --> TermRecordConfig : uses
    TermDictionarySearcher <|.. SQLiteTermDictionarySearcher : implements
    SQLiteTermDictionarySearcher --> KeywordStemmer : uses
    ModTermStore <|.. SQLiteModTermStore : implements
    TermTranslatorImpl ..> TermTranslationRequest : creates
    TermTranslatorImpl ..> TermTranslationResult : creates
    ModTermStore ..> TermTranslationResult : stores
```

## アーキテクチャの補足：2フェーズモデル (Propose/Save)
本スライスはバッチAPIや長時間実行ジョブに対応するため、**「プロンプト生成(ProposeJobs)」**と**「結果保存(SaveResults)」**の2フェーズに分割されている。
- **Phase 1 (Propose)**: 入力データを解析し、辞書検索の結果に基づいたコンテキストを含むLLMリクエストを生成する。既訳がある場合はLLMを介さず即時結果として返す。
- **Phase 2 (Save)**: JobQueue等を通じて取得されたLLMのレスポンス群を受け取り、パースしてMod用語DBに永続化する。

スライス自身はLLM Clientを直接呼び出さず、呼び出し元のオーケストレーター（ProcessManager）がリクエストの実行順序や並列度を制御する。
