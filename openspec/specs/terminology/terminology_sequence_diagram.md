# 用語翻訳・Mod用語DB保存 シーケンス図

## 1. 用語翻訳メインフロー（2フェーズモデル）

2フェーズモデルでは、スライスは「ジョブ提案 (Phase 1)」と「結果保存 (Phase 2)」の2つの独立した Contract メソッドとして呼び出される。LLMとの通信は外部（Pipeline / JobQueue）が担当する。

### Phase 1: 用語翻訳ジョブの提案 (Propose)

```mermaid
sequenceDiagram
    autonumber
    participant PM as Pipeline
    participant TT as Terminology<br/>(Interface)
    participant RB as TermRequestBuilder
    participant DS as TermDictionarySearcher
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    PM->>TT: ProposeJobs(ctx, input)

    Note over TT: Step 1: リクエスト生成（NPC FULL+SHRTペアリング含む）
    TT->>RB: BuildRequests(ctx, input)
    Note over RB: 同一EditorIDのNPC_:FULLとNPC_:SHRTを<br/>1つのTermTranslationRequestにペアリング
    RB-->>TT: []TermTranslationRequest

    Note over TT: Step 2: 辞書検索
    TT->>DS: SearchBatch(ctx, sourceTexts)
    DS->>DB: SELECT FROM dictionaries (公式DLC辞書)
    DB-->>DS: 既訳用語
    DS-->>TT: map[string][]ReferenceTerm

    Note over TT: Step 3: 強制翻訳の解決 & ジョブ構築
    loop 各リクエスト
        alt 既訳完全一致あり (Forced Translation)
            TT->>TT: 既訳を即時結果としてキャッシュ
        else 既訳なし
            TT->>TT: LLMプロンプト構築
            TT->>TT: []llm_client.Request に追加
        end
    end

    TT-->>PM: ProposeOutput (Requests, PreCalculatedResults)
```

### Phase 2: 用語翻訳結果の保存 (Save)

```mermaid
sequenceDiagram
    autonumber
    participant PM as Pipeline
    participant TT as Terminology<br/>(Interface)
    participant MS as ModTermStore<br/>(Interface)
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    PM->>TT: SaveResults(ctx, responses)

    Note over TT: Step 1: レスポンスのパース・バリデーション
    loop 各レスポンス
        alt Success == true
            TT->>TT: "TL: |...|" フォーマット抽出
        else Success == false
            TT->>TT: エラーログ記録 (Skip)
        end
    end

    Note over TT: Step 2: Mod用語DB保存（UPSERT）
    TT->>MS: InitSchema(ctx)
    MS->>DB: CREATE TABLE IF NOT EXISTS mod_terms ...
    TT->>MS: SaveTerms(ctx, results)
    MS->>DB: BEGIN TRANSACTION
    MS->>DB: INSERT OR REPLACE INTO mod_terms ...
    MS->>DB: COMMIT
    MS-->>TT: Success

    TT-->>PM: error (nil if success)
```

## 2. 辞書検索の詳細フロー（貪欲部分一致）

```mermaid
sequenceDiagram
    autonumber
    participant TT as Terminology
    participant DS as TermDictionarySearcher
    participant GM as GreedyLongestMatcher
    participant DictDB as 辞書DB<br/>(公式DLC)
    participant KS as KeywordStemmer
    participant NPCFTS as npc_terms_fts<br/>(NPC専用FTS5)

    TT->>DS: SearchBatch(ctx, sourceTexts)

    loop 各ソーステキスト
        Note over DS: Step 1: ソーステキスト全文の完全一致（最優先）
        DS->>DictDB: SELECT * FROM dictionaries WHERE source = ?
        DictDB-->>DS: 完全一致結果

        Note over DS: Step 2: キーワード抽出
        DS->>DS: extractSearchKeywords(text)

        Note over DS: Step 3: キーワード完全一致検索
        DS->>DictDB: SELECT * FROM dictionaries WHERE source IN (keywords)
        DictDB-->>DS: キーワード一致結果

        Note over DS: Step 3.5: ステミングによるフォールバック検索
        DS->>KS: StripPossessive + Stem(未ヒットキーワード)
        KS-->>DS: stemmed keywords
        DS->>DictDB: 辞書sourceもステム化して比較
        DictDB-->>DS: ステム一致結果（低優先）

        Note over DS: Step 4: 貪欲最長一致フィルタリング
        DS->>GM: Filter(sourceText, candidates)
        GM-->>DS: フィルタリング済み候補

        Note over DS: Step 5: NPC名の貪欲部分一致
        DS->>DS: 未消費キーワードに対してステミング検索
        DS->>NPCFTS: SELECT * FROM npc_terms_fts WHERE MATCH ? (原形+ステム)
        NPCFTS-->>DS: NPC部分一致結果
    end

    DS-->>TT: map[string][]ReferenceTerm
```
