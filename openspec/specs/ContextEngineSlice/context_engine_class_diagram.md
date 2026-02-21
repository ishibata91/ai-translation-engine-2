# コンテキストエンジン クラス図

## クラス構成

```mermaid
classDiagram
    class ContextEngine {
        <<interface>>
        +BuildTranslationRequests(data *ExtractedData, config ContextEngineConfig) ([]TranslationRequest, error)
    }

    class ContextEngineImpl {
        -toneResolver ToneResolver
        -personaLookup PersonaLookup
        -termLookup TermLookup
        -summaryLookup SummaryLookup
        +BuildTranslationRequests(data, config)
        -buildDialogueRequests(groups []DialogueGroup, npcs map~string~NPC, config ContextEngineConfig) []TranslationRequest
        -buildQuestRequests(quests []Quest, config ContextEngineConfig) []TranslationRequest
        -buildItemRequests(items []Item, config ContextEngineConfig) []TranslationRequest
        -buildMagicRequests(magics []Magic, config ContextEngineConfig) []TranslationRequest
        -buildMessageRequests(messages []Message, config ContextEngineConfig) []TranslationRequest
        -resolveSpeaker(speakerID *string, npcs map~string~NPC) *SpeakerProfile
        -getTopicName(response DialogueResponse, group DialogueGroup) string
        -containsJapanese(text string) bool
    }

    class ToneResolver {
        <<interface>>
        +Resolve(race string, voiceType string, sex string) string
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
        -voiceToneMap map~string~string
        +Resolve(race, voiceType, sex) string
    }

    class SQLitePersonaLookup {
        -db *sql.DB
        +FindBySpeakerID(speakerID) (*string, error)
    }

    class SQLiteTermLookup {
        -dbs []*sql.DB
        -stemmer Stemmer
        +Search(sourceText) ([]ReferenceTerm, *string, error)
    }

    class SQLiteSummaryLookup {
        -db *sql.DB
        +FindDialogueSummary(groupID) (*string, error)
        +FindQuestSummary(questID) (*string, error)
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
        +SourceFile string
    }

    TranslationRequest --> TranslationContext
    TranslationRequest --> ReferenceTerm
    TranslationContext --> SpeakerProfile
```

## 依存関係

- `ContextEngineImpl` → `ToneResolver`: 口調指示文の生成
- `ContextEngineImpl` → `PersonaLookup`: NPCペルソナの参照
- `ContextEngineImpl` → `TermLookup`: 参照用語の検索・強制翻訳判定
- `ContextEngineImpl` → `SummaryLookup`: 会話要約・クエスト要約の参照
- `ConfigToneResolver` → Config Store: 種族・ボイスタイプの口調マッピング取得
- `SQLitePersonaLookup` → `*sql.DB` (DI): ペルソナDBへの読み取りアクセス
- `SQLiteTermLookup` → `[]*sql.DB` (DI): 辞書DB・Mod用語DBへの読み取りアクセス
- `SQLiteSummaryLookup` → `*sql.DB` (DI): 要約キャッシュDBへの読み取りアクセス
- Process Manager → `ContextEngine`: 翻訳リクエスト構築の起動
