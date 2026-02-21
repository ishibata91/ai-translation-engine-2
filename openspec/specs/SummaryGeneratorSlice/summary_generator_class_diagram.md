# 要約ジェネレータ クラス図

## クラス構成

```
┌─────────────────────────────────────────┐
│         <<interface>>                   │
│        SummaryGenerator                 │
├─────────────────────────────────────────┤
│ + GenerateDialogueSummaries(            │
│     ctx context.Context,                │
│     groups []DialogueGroupInput,        │
│     progress func(done, total int),     │
│   ) ([]SummaryResult, error)            │
│                                         │
│ + GenerateQuestSummaries(               │
│     ctx context.Context,                │
│     quests []QuestInput,                │
│     progress func(done, total int),     │
│   ) ([]SummaryResult, error)            │
└─────────────────────────────────────────┘
                    ▲
                    │ implements
┌─────────────────────────────────────────┐
│        SummaryGeneratorImpl             │
├─────────────────────────────────────────┤
│ - llmClient    LLMClient               │
│ - store        SummaryStore             │
│ - hasher       CacheKeyHasher           │
│ - concurrency  int                      │
│ - maxTokens    int                      │
│ - temperature  float64                  │
├─────────────────────────────────────────┤
│ + GenerateDialogueSummaries(...)        │
│ + GenerateQuestSummaries(...)           │
│ - summarizeSingle(ctx, input) string    │
│ - buildDialoguePrompt(lines) string     │
│ - buildQuestPrompt(stages) string       │
└─────────────────────────────────────────┘
          │              │
          │ uses         │ uses
          ▼              ▼
┌──────────────┐  ┌─────────────────────────────┐
│  LLMClient   │  │     <<interface>>            │
│ (共通infra)  │  │      SummaryStore            │
└──────────────┘  ├─────────────────────────────┤
                  │ + InitTable(ctx) error       │
                  │ + Get(ctx, cacheKey)          │
                  │     (*SummaryRecord, error)   │
                  │ + Upsert(ctx, record) error   │
                  │ + GetByRecordID(ctx,          │
                  │     recordID, summaryType)    │
                  │     (*SummaryRecord, error)   │
                  │ + Close() error               │
                  └─────────────────────────────┘
                              ▲
                              │ implements
                  ┌─────────────────────────────┐
                  │    SQLiteSummaryStore        │
                  ├─────────────────────────────┤
                  │ - db        *sql.DB          │
                  │ - dbPath    string           │
                  ├─────────────────────────────┤
                  │ + NewSQLiteSummaryStore(     │
                  │     cacheDir string,         │
                  │     sourcePlugin string,     │
                  │   ) (*SQLiteSummaryStore,    │
                  │      error)                  │
                  │ + InitTable(ctx) error       │
                  │ + Get(ctx, cacheKey)          │
                  │     (*SummaryRecord, error)   │
                  │ + Upsert(ctx, record) error   │
                  │ + GetByRecordID(ctx,          │
                  │     recordID, summaryType)    │
                  │     (*SummaryRecord, error)   │
                  │ + Close() error               │
                  └─────────────────────────────┘

┌─────────────────────────────────────────┐
│          CacheKeyHasher                 │
├─────────────────────────────────────────┤
│ + BuildCacheKey(                        │
│     recordID string,                    │
│     lines []string,                     │
│   ) string                              │
└─────────────────────────────────────────┘
```

## DTO定義

```
┌─────────────────────────────────────────┐
│         DialogueGroupInput              │
├─────────────────────────────────────────┤
│ + GroupID       string                  │
│ + PlayerText    *string                 │
│ + Lines         []string                │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│            QuestInput                   │
├─────────────────────────────────────────┤
│ + QuestID       string                  │
│ + StageTexts    []string                │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│          SummaryResult                  │
├─────────────────────────────────────────┤
│ + RecordID      string                  │
│ + SummaryType   string                  │
│ + SummaryText   string                  │
│ + CacheHit      bool                    │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│          SummaryRecord                  │
├─────────────────────────────────────────┤
│ + ID             int64                  │
│ + RecordID       string                 │
│ + SummaryType    string                 │
│ + CacheKey       string                 │
│ + InputHash      string                 │
│ + SummaryText    string                 │
│ + InputLineCount int                    │
│ + CreatedAt      time.Time              │
│ + UpdatedAt      time.Time              │
└─────────────────────────────────────────┘
```

## ソースファイル単位キャッシュの構成

```
cache_dir/
├── Skyrim.esm_summary_cache.db        ← Skyrim.esm 用
│   └── summaries テーブル
├── Dawnguard.esm_summary_cache.db     ← Dawnguard.esm 用
│   └── summaries テーブル
└── MyMod.esp_summary_cache.db         ← MyMod.esp 用
    └── summaries テーブル
```

- `NewSQLiteSummaryStore(cacheDir, sourcePlugin)` がソースファイル名からDBファイルパスを決定し、接続を確立する。
- 命名規則: `{sourcePlugin}_summary_cache.db`
- 各DBファイルは独立しており、Mod単位での削除・再生成・配布が容易。

## 依存関係

- `SummaryGeneratorImpl` → `LLMClient` (共通インフラ): LLM呼び出し
- `SummaryGeneratorImpl` → `SummaryStore`: キャッシュの読み書き
- `SummaryGeneratorImpl` → `CacheKeyHasher`: キャッシュキー生成
- `SQLiteSummaryStore` → `*sql.DB` (内部生成): ソースファイル単位のDB接続
- Process Manager → `SummaryGenerator`: 要約生成の起動
