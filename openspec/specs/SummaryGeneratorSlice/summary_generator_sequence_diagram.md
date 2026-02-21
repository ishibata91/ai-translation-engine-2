# 要約ジェネレータ シーケンス図

## 0. ソースファイル単位のStore初期化フロー

```
ProcessManager                          SQLiteSummaryStore
     │                                         │
     │  NewSQLiteSummaryStore(cacheDir, "Skyrim.esm")
     │────────────────────────────────────────>│
     │                                         │── open "cacheDir/Skyrim.esm_summary_cache.db"
     │                                         │
     │  *SQLiteSummaryStore, nil               │
     │<────────────────────────────────────────│
     │                                         │
     │  InitTable(ctx)                         │
     │────────────────────────────────────────>│
     │                                         │── CREATE TABLE IF NOT EXISTS summaries ...
     │                                         │── CREATE INDEX IF NOT EXISTS ...
     │  nil                                    │
     │<────────────────────────────────────────│
     │                                         │
     │  (ソースファイルごとに繰り返し)          │
```

## 1. 会話要約生成フロー

```
ProcessManager          SummaryGeneratorImpl       CacheKeyHasher       SummaryStore(SQLite)       LLMClient
     │                         │                        │                      │                      │
     │ GenerateDialogueSummaries(ctx, groups, progress)  │                      │                      │
     │────────────────────────>│                        │                      │                      │
     │                         │                        │                      │                      │
     │                         │── for each group ──────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │                         │  BuildCacheKey(groupID, lines)                │                      │
     │                         │───────────────────────>│                      │                      │
     │                         │  cacheKey              │                      │                      │
     │                         │<──────────────────────│                      │                      │
     │                         │                        │                      │                      │
     │                         │  Get(ctx, cacheKey)    │                      │                      │
     │                         │───────────────────────────────────────────────>│                      │
     │                         │  *SummaryRecord / nil  │                      │                      │
     │                         │<─────────────────────────────────────────────│                      │
     │                         │                        │                      │                      │
     │                         │── [cache miss] ────────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │                         │  buildDialoguePrompt(lines)                   │                      │
     │                         │──(internal)            │                      │                      │
     │                         │                        │                      │                      │
     │                         │  Complete(LLMRequest{system, user, maxTokens=200, temp=0.3})         │
     │                         │─────────────────────────────────────────────────────────────────────>│
     │                         │  LLMResponse{content}  │                      │                      │
     │                         │<────────────────────────────────────────────────────────────────────│
     │                         │                        │                      │                      │
     │                         │  Upsert(ctx, SummaryRecord)                   │                      │
     │                         │───────────────────────────────────────────────>│                      │
     │                         │  ok                    │                      │                      │
     │                         │<─────────────────────────────────────────────│                      │
     │                         │                        │                      │                      │
     │                         │── end [cache miss] ────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │  progress(done, total)  │                        │                      │                      │
     │<────────────────────────│                        │                      │                      │
     │                         │                        │                      │                      │
     │                         │── end for each ────────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │  []SummaryResult        │                        │                      │                      │
     │<────────────────────────│                        │                      │                      │
```

## 2. クエスト要約生成フロー（累積的処理）

```
ProcessManager          SummaryGeneratorImpl       CacheKeyHasher       SummaryStore(SQLite)       LLMClient
     │                         │                        │                      │                      │
     │ GenerateQuestSummaries(ctx, quests, progress)     │                      │                      │
     │────────────────────────>│                        │                      │                      │
     │                         │                        │                      │                      │
     │                         │── for each quest ──────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │                         │  sort stages by Index (ascending)             │                      │
     │                         │──(internal)            │                      │                      │
     │                         │                        │                      │                      │
     │                         │  BuildCacheKey(questID, allStageTexts)        │                      │
     │                         │───────────────────────>│                      │                      │
     │                         │  cacheKey              │                      │                      │
     │                         │<──────────────────────│                      │                      │
     │                         │                        │                      │                      │
     │                         │  Get(ctx, cacheKey)    │                      │                      │
     │                         │───────────────────────────────────────────────>│                      │
     │                         │  *SummaryRecord / nil  │                      │                      │
     │                         │<─────────────────────────────────────────────│                      │
     │                         │                        │                      │                      │
     │                         │── [cache miss] ────────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │                         │  buildQuestPrompt(stageTexts)                 │                      │
     │                         │──(internal)            │                      │                      │
     │                         │                        │                      │                      │
     │                         │  Complete(LLMRequest{system, user, maxTokens=200, temp=0.3})         │
     │                         │─────────────────────────────────────────────────────────────────────>│
     │                         │  LLMResponse{content}  │                      │                      │
     │                         │<────────────────────────────────────────────────────────────────────│
     │                         │                        │                      │                      │
     │                         │  Upsert(ctx, SummaryRecord)                   │                      │
     │                         │───────────────────────────────────────────────>│                      │
     │                         │  ok                    │                      │                      │
     │                         │<─────────────────────────────────────────────│                      │
     │                         │                        │                      │                      │
     │                         │── end [cache miss] ────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │  progress(done, total)  │                        │                      │                      │
     │<────────────────────────│                        │                      │                      │
     │                         │                        │                      │                      │
     │                         │── end for each ────────────────────────────────────────────────────── │
     │                         │                        │                      │                      │
     │  []SummaryResult        │                        │                      │                      │
     │<────────────────────────│                        │                      │                      │
```

## 3. Pass 2 参照フロー

```
TranslatorSlice         SummaryStore(SQLite)
     │                         │
     │  (該当ソースファイルのSummaryStoreを使用)
     │                         │
     │  GetByRecordID(ctx, dialogueGroupID, "dialogue")
     │────────────────────────>│
     │  *SummaryRecord / nil   │
     │<────────────────────────│
     │                         │
     │  GetByRecordID(ctx, questID, "quest")
     │────────────────────────>│
     │  *SummaryRecord / nil   │
     │<────────────────────────│
     │                         │
     │  (要約をLLMプロンプトのコンテキストに挿入)
     │──(internal)             │
```

## 4. 並列実行モデル

```
SummaryGeneratorImpl
     │
     │── Goroutine Pool (semaphore: concurrency=10) ──
     │                                                 │
     │  ┌─ goroutine 1: summarize(group[0]) ─────────┐│
     │  │  cache check → [miss] → LLM call → upsert  ││
     │  └─────────────────────────────────────────────┘│
     │  ┌─ goroutine 2: summarize(group[1]) ─────────┐│
     │  │  cache check → [hit] → return cached        ││
     │  └─────────────────────────────────────────────┘│
     │  ┌─ goroutine 3: summarize(group[2]) ─────────┐│
     │  │  cache check → [miss] → LLM call → upsert  ││
     │  └─────────────────────────────────────────────┘│
     │  ...                                            │
     │─────────────────────────────────────────────────│
     │
     │  collect results, notify progress
```
