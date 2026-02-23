# NPCペルソナ生成 シーケンス図

## 1. ペルソナ生成メインフロー（2フェーズモデル）

2フェーズモデルでは、スライスは「ジョブ提案 (Phase 1)」と「結果保存 (Phase 2)」の2つの独立した Contract メソッドとして呼び出される。

### Phase 1: ペルソナ生成ジョブの提案 (Propose)

```mermaid
sequenceDiagram
    autonumber
    participant PM as Process Manager
    participant PG as NPCPersonaGenerator<br/>(Interface)
    participant DC as DialogueCollector
    participant CE as ContextEvaluator
    participant PB as PersonaPromptBuilder

    PM->>PG: ProposeJobs(ctx, input)

    Note over PG: Step 1: NPC会話データ収集
    PG->>DC: CollectByNPC(ctx, input)
    DC-->>PG: []NPCDialogueData

    Note over PG: Step 2: コンテキスト長評価 & トークン事前計算
    loop 各NPCDialogueData
        PG->>CE: Evaluate(dialogueData, config)
        CE-->>PG: TokenEstimation, []DialogueEntry（調整済み）
    end

    Note over PG: Step 3: ジョブ構築
    loop 各評価済みNPC
        PG->>PB: BuildSystemPrompt()
        PB-->>PG: systemPrompt
        PG->>PB: BuildUserPrompt(dialogueData)
        PB-->>PG: userPrompt
        PG->>PG: []llm_client.Request に追加
    end

    PG-->>PM: ProposeOutput (Requests)
```

### Phase 2: ペルソナ生成結果の保存 (Save)

```mermaid
sequenceDiagram
    autonumber
    participant PM as Process Manager
    participant PG as NPCPersonaGenerator<br/>(Interface)
    participant PS as PersonaStore<br/>(Interface)
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    PM->>PG: SaveResults(ctx, responses)

    Note over PG: Step 1: レスポンスのパース・バリデーション
    loop 各レスポンス
        alt Success == true
            PG->>PG: 性格・口調・背景セクション抽出
        else Success == false
            PG->>PG: エラーログ記録 (Skip)
        end
    end

    Note over PG: Step 2: ペルソナDB保存（UPSERT）
    PG->>PS: InitSchema(ctx)
    PS->>DB: CREATE TABLE IF NOT EXISTS npc_personas ...
    PG->>PS: SavePersona(ctx, results)
    PS->>DB: INSERT OR REPLACE INTO npc_personas ...
    MS-->>PG: Success

    PG-->>PM: error (nil if success)
```

## 2. Pass 2 でのペルソナ参照フロー

```mermaid
sequenceDiagram
    autonumber
    participant MT as MainTranslator<br/>(Pass 2)
    participant PS as PersonaStore
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    MT->>PS: GetPersona(ctx, speakerID)
    PS->>DB: SELECT persona_text FROM npc_personas WHERE speaker_id = ?
    DB-->>PS: 結果

    alt ペルソナが存在する
        PS-->>MT: persona_text
        Note over MT: システムプロンプトの話者プロファイルに挿入
    else ペルソナが存在しない
        PS-->>MT: "" (空文字)
        Note over MT: 従来の口調推定にフォールバック
    end
```
