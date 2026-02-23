# コンテキストエンジン シーケンス図

## 1. 全体フロー（BuildTranslationRequests）

JobQueue連携モデルでは、ContextEngineは入力データを解析し、コンテキスト情報を付与した `TranslationRequest` 群を構築して返す。実際の翻訳実行は Pass 2 Translator Slice が担当する。

```mermaid
sequenceDiagram
    autonumber
    participant PM as ProcessManager
    participant CE as ContextEngineImpl
    participant TR as ToneResolver
    participant PL as PersonaLookup
    participant TL as TermLookup
    participant SL as SummaryLookup

    PM->>CE: BuildTranslationRequests(ctx, input, config)

    Note over CE: Step 1: 会話リクエスト構築 & ツリー解析
    loop 各DialogueGroup
        CE->>SL: FindDialogueSummary(groupID)
        CE->>CE: resolveSpeaker(npcs)
        CE->>TR: Resolve(race, class, voice, sex)
        CE->>PL: FindBySpeakerID(speakerID)
        CE->>TL: Search(text)
        Note over CE: → TranslationRequest (PreviousLineトラッキング含む)
    end

    Note over CE: Step 2: クエストリクエスト構築
    loop 各Quest
        CE->>SL: FindQuestSummary(questID)
        CE->>TL: Search(text)
        Note over CE: → TranslationRequest (QUST CNAM/NNAM)
    end

    Note over CE: Step 3: アイテム・書籍・魔法・メッセージ構築
    loop 各Item/Magic/Message
        CE->>TL: Search(text)
        Note over CE: → TranslationRequest (書籍チャンク分割含む)
    end

    CE-->>PM: []TranslationRequest
```

## 2. 参照用語検索と強制翻訳判定（TermLookup）

```mermaid
sequenceDiagram
    autonumber
    participant CE as ContextEngineImpl
    participant TL as TermLookup
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    CE->>TL: Search(sourceText)

    Note over TL: Step 1: 完全一致チェック
    TL->>DB: SELECT translated_ja FROM dictionaries/mod_terms WHERE source = ?
    alt 完全一致あり
        DB-->>TL: 既訳
        TL-->>CE: ReferenceTerms, *forcedTranslation
    else 一致なし
        Note over TL: Step 2: キーワード貪欲部分一致検索
        TL->>DB: キーワード抽出・ステミング・検索
        DB-->>TL: ヒット用語群
        TL-->>CE: []ReferenceTerm, nil
    end
```
