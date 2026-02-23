# 要約ジェネレータ クラス図

## クラス構成

```mermaid
classDiagram
    class SummaryGeneratorInput {
        +[]DialogueGroupInput DialogueGroups
        +[]QuestInput Quests
    }

    class ProposeOutput {
        +[]llm_client.Request Requests
        +[]SummaryResult PreCalculatedResults
    }

    class SummaryGenerator {
        <<interface>>
        +ProposeJobs(ctx context.Context, input SummaryGeneratorInput) (ProposeOutput, error)
        +SaveResults(ctx context.Context, responses []llm_client.Response) error
        +GetSummary(ctx context.Context, recordID string, summaryType string) (*SummaryResult, error)
    }

    class SummaryGeneratorImpl {
        -store SummaryStore
        -hasher CacheKeyHasher
        -promptBuilder SummaryPromptBuilder
        +ProposeJobs(...)
        +SaveResults(...)
        +GetSummary(...)
    }

    class SummaryStore {
        <<interface>>
        +InitTable(ctx) error
        +Get(ctx, cacheKey) (*SummaryRecord, error)
        +Upsert(ctx, record) error
        +GetByRecordID(ctx, recordID, summaryType) (*SummaryRecord, error)
        +Close() error
    }

    class SQLiteSummaryStore {
        -db *sql.DB
        +InitTable(ctx) error
        +Get(ctx, cacheKey) (*SummaryRecord, error)
        +Upsert(ctx, record) error
        +GetByRecordID(ctx, recordID, summaryType) (*SummaryRecord, error)
    }

    class CacheKeyHasher {
        +BuildCacheKey(recordID string, lines []string) string
    }

    class SummaryPromptBuilder {
        +BuildDialoguePrompt(lines []string) string
        +BuildQuestPrompt(stages []string) string
    }

    SummaryGenerator <|.. SummaryGeneratorImpl : implements
    SummaryGeneratorImpl --> SummaryStore : uses
    SummaryGeneratorImpl --> CacheKeyHasher : uses
    SummaryGeneratorImpl --> SummaryPromptBuilder : uses
    SummaryStore <|.. SQLiteSummaryStore : implements
```

## DTO定義

```mermaid
classDiagram
    class DialogueGroupInput {
        +GroupID string
        +PlayerText *string
        +Lines []string
    }

    class QuestInput {
        +QuestID string
        +StageTexts []string
    }

    class SummaryResult {
        +RecordID string
        +SummaryType string
        +SummaryText string
        +CacheHit bool
    }

    class SummaryRecord {
        +ID int64
        +RecordID string
        +SummaryType string
        +CacheKey string
        +InputHash string
        +SummaryText string
        +InputLineCount int
        +CreatedAt time.Time
        +UpdatedAt time.Time
    }
```

## アーキテクチャの補足：2フェーズモデル (Propose/Save)
本スライスはバッチAPIや長時間実行ジョブに対応するため、**「プロンプト生成(ProposeJobs)」**と**「結果保存(SaveResults)」**の2フェーズに分割されている。
- **Phase 1 (Propose)**: 入力データを解析し、キャッシュヒット判定を行う。既訳がない場合はLLMプロンプトを生成してリクエスト群として返す。既訳がある場合は即時結果として返す。
- **Phase 2 (Save)**: JobQueue等を通じて取得されたLLMのレスポンス群を受け取り、パースしてSQLiteキャッシュに永続化する。

スライス自身はLLM Clientを直接呼び出さず、呼び出し元のオーケストレーター（ProcessManager）が通信を制御する。
