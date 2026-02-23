# 本文翻訳 シーケンス図

## 1. 本文翻訳メインフロー（2フェーズモデル）

2フェーズモデルでは、スライスは「ジョブ提案 (Phase 1)」と「結果保存 (Phase 2)」の2つの独立した Contract メソッドとして呼び出される。

### Phase 1: 翻訳ジョブの提案 (Propose)

```mermaid
sequenceDiagram
    autonumber
    participant PM as ProcessManager
    participant BT as BatchTranslator<br/>(Interface)
    participant RL as ResumeLoader
    participant TP as TagProcessor
    participant PB as PromptBuilder

    PM->>BT: ProposeJobs(ctx, requests, config)

    Note over BT: Step 1: 差分更新チェック (Resume)
    BT->>RL: LoadCachedResults(config)
    RL-->>BT: map[key]Result

    loop for each request
        alt cache hit
            BT->>BT: キャッシュ済みの結果として即時リターン用リストに追加
        else cache miss
            Note over BT: Step 2: HTMLタグ保護 & プロンプト構築
            BT->>TP: Preprocess(sourceText)
            TP-->>BT: processedText, tagMap
            BT->>PB: Build(request)
            PB-->>BT: systemPrompt, userPrompt
            BT->>BT: []llm_client.Request に追加
        end
    end

    BT-->>PM: ProposeOutput (Requests, PreCalculatedResults)
```

### Phase 2: 翻訳結果の保存 (Save)

```mermaid
sequenceDiagram
    autonumber
    participant PM as ProcessManager
    participant BT as BatchTranslator<br/>(Interface)
    participant TP as TagProcessor
    participant TV as TranslationVerifier
    participant RW as ResultWriter

    PM->>BT: SaveResults(ctx, responses)

    loop for each response
        alt Success == true
            Note over BT: Step 1: タグ復元 & バリデーション
            BT->>TP: Postprocess(translated, tagMap)
            TP-->>BT: restoredText
            BT->>TP: Validate(restoredText, tagMap)
            BT->>TV: Verify(source, translated, tagMap)
            
            Note over BT: Step 2: 逐次保存
            BT->>RW: Write(result)
            RW->>RW: JSONファイル更新
        else Success == false
            BT->>BT: エラーログ記録 (Skip / Notify Retry)
        end
    end

    BT->>RW: Flush()
    BT-->>PM: error (nil if success)
```

## 2. 書籍長文分割翻訳フロー（Propose時）

```mermaid
sequenceDiagram
    autonumber
    participant BT as BatchTranslator
    participant BC as BookChunker

    Note over BT: ProposeJobs 内で BOOK DESC を検知
    BT->>BC: Chunk(text, maxTokensPerChunk)
    BC-->>BT: []chunks
    loop 各チャンク
        BT->>BT: 個別の llm_client.Request を生成
    end
```
