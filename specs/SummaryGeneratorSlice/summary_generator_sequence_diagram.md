# 要約ジェネレータ シーケンス図

## 1. 要約生成フロー（2フェーズモデル）

2フェーズモデルでは、スライスは「ジョブ提案 (Phase 1)」と「結果保存 (Phase 2)」の2つの独立した Contract メソッドとして呼び出される。

### Phase 1: 要約ジョブの提案 (Propose)

```mermaid
sequenceDiagram
    autonumber
    participant PM as ProcessManager
    participant SG as SummaryGenerator<br/>(Interface)
    participant CK as CacheKeyHasher
    participant SS as SummaryStore (SQLite)

    PM->>SG: ProposeJobs(ctx, input)

    loop for each group/quest
        SG->>CK: BuildCacheKey(id, texts)
        CK-->>SG: cacheKey

        SG->>SS: Get(ctx, cacheKey)
        SS-->>SG: *SummaryRecord / nil

        alt cache hit
            SG->>SG: キャッシュ済みの結果として即時リターン用リストに追加
        else cache miss
            SG->>SG: LLMプロンプト構築
            SG->>SG: []llm_client.Request に追加
        end
    end

    SG-->>PM: ProposeOutput (Requests, PreCalculatedResults)
```

### Phase 2: 要約結果の保存 (Save)

```mermaid
sequenceDiagram
    autonumber
    participant PM as ProcessManager
    participant SG as SummaryGenerator<br/>(Interface)
    participant SS as SummaryStore (SQLite)

    PM->>SG: SaveResults(ctx, responses)

    loop for each response
        alt Success == true
            SG->>SG: 要約テキスト抽出・バリデーション
            SG->>SS: Upsert(ctx, SummaryRecord)
            SS-->>SG: ok
        else Success == false
            SG->>SG: エラーログ記録 (Skip)
        end
    end

    SG-->>PM: error (nil if success)
```

## 2. Pass 2 参照フロー

```mermaid
sequenceDiagram
    participant TS as TranslatorSlice
    participant SS as SummaryStore (SQLite)

    Note over TS,SS: 該当ソースファイルのSummaryStoreを使用

    TS->>SS: GetByRecordID(ctx, recordID, summaryType)
    SS-->>TS: *SummaryRecord / nil

    Note over TS: 要約をLLMプロンプトのコンテキストに挿入
```
