# コンテキストエンジン クラス図

## クラス構成

```mermaid
classDiagram
    class ContextEngineInput {
        +[]NPC NPCs
        +[]DialogueGroup DialogueGroups
        +[]DialogueResponse DialogueResponses
        +[]Quest Quests
        +[]Item Items
        +[]Magic Magics
        +[]Message Messages
    }

    class ContextEngine {
        <<interface>>
        +BuildTranslationRequests(ctx context.Context, input ContextEngineInput, config ContextEngineConfig) ([]TranslationRequest, error)
    }

    class ContextEngineImpl {
        -toneResolver ToneResolver
        -personaLookup PersonaLookup
        -termLookup TermLookup
        -summaryLookup SummaryLookup
        +BuildTranslationRequests(...)
    }

    class ToneResolver {
        <<interface>>
        +Resolve(race string, class string, voiceType string, sex string) string
    }

    class PersonaLookup {
        <<interface>>
        +FindBySpeakerID(speakerID string) (*string, error)
    }

    class TermLookup {
        <<interface>>
        +Search(sourceText string) ([]ReferenceTerm, *string, error)
    }

    class SummaryLookup {
        <<interface>>
        +FindDialogueSummary(groupID string) (*string, error)
        +FindQuestSummary(questID string) (*string, error)
    }

    class ConfigToneResolver {
        -raceToneMap map~string~string
        -classToneMap map~string~string
        -voiceToneMap map~string~string
        +Resolve(...)
    }

    class SQLitePersonaLookup {
        -db *sql.DB
        +FindBySpeakerID(...)
    }

    class SQLiteTermLookup {
        -dbs []*sql.DB
        +Search(...)
    }

    class SQLiteSummaryLookup {
        -db *sql.DB
        +FindDialogueSummary(...)
        +FindQuestSummary(...)
    }

    ContextEngine <|.. ContextEngineImpl : implements
    ContextEngineImpl --> ToneResolver : uses
    ContextEngineImpl --> PersonaLookup : uses
    ContextEngineImpl --> TermLookup : uses
    ContextEngineImpl --> SummaryLookup : uses
    ToneResolver <|.. ConfigToneResolver : implements
    PersonaLookup <|.. SQLitePersonaLookup : implements
    TermLookup <|.. SQLiteTermLookup : implements
    SummaryLookup <|.. SQLiteSummaryLookup : implements
```

## DTO定義

```mermaid
classDiagram
    class TranslationRequest {
        +ID string
        +RecordType string
        +SourceText string
        +Context TranslationContext
        +Index *int
        +ReferenceTerms []ReferenceTerm
        +EditorID *string
        +ForcedTranslation *string
        +SourcePlugin string
        +SourceFile string
        +MaxTokens *int
    }

    class TranslationContext {
        +PreviousLine *string
        +Speaker *SpeakerProfile
        +TopicName *string
        +QuestName *string
        +QuestSummary *string
        +DialogueSummary *string
        +ItemTypeHint *string
        +ModDescription *string
        +PlayerTone *string
    }

    class SpeakerProfile {
        +Name string
        +Gender string
        +Race string
        +VoiceType string
        +ToneInstruction string
        +PersonaText *string
    }

    class ReferenceTerm {
        +OriginalEN string
        +OriginalJA string
    }

    class ContextEngineConfig {
        +ModDescription string
        +PlayerTone string
    }

    TranslationRequest --> TranslationContext
    TranslationRequest --> ReferenceTerm
    TranslationContext --> SpeakerProfile
```

## 依存関係
- `ContextEngineImpl` → `ToneResolver`: 属性からの口調指示生成
- `ContextEngineImpl` → `PersonaLookup`: 保存済みNPCペルソナの検索 (Phase 1)
- `ContextEngineImpl` → `TermLookup`: 既訳辞書からの用語抽出・強制翻訳判定 (Phase 1)
- `ContextEngineImpl` → `SummaryLookup`: 保存済み要約の検索 (Phase 1)
- Process Manager → `ContextEngine`: 翻訳ジョブ（リクエスト群）の構築
