# NPCペルソナ生成 クラス図

```mermaid
classDiagram
    class PersonaGenInput {
        +[]NPC NPCs
        +[]DialogueGroup DialogueGroups
        +[]DialogueResponse DialogueResponses
    }

    class ProposeOutput {
        +[]llm_client.Request Requests
    }

    class NPCDialogueData {
        +String SpeakerID
        +String EditorID
        +String NPCName
        +String Race
        +String Sex
        +String VoiceType
        +[]DialogueEntry Dialogues
    }

    class DialogueEntry {
        +String Text
        +String EnglishText
        +String QuestID
        +Bool IsServicesBranch
        +Int Order
    }

    class PersonaResult {
        +String SpeakerID
        +String EditorID
        +String NPCName
        +String Race
        +String Sex
        +String VoiceType
        +String PersonaText
        +Int DialogueCount
        +Int EstimatedTokens
        +String SourcePlugin
        +String Status
        +String ErrorMessage
    }

    class NPCPersonaGenerator {
        <<Interface>>
        +ProposeJobs(ctx context.Context, input PersonaGenInput) (ProposeOutput, error)
        +SaveResults(ctx context.Context, responses []llm_client.Response) error
        +GetPersona(ctx context.Context, speakerID string) (string, error)
    }

    class NPCPersonaGeneratorImpl {
        -DialogueCollector collector
        -ContextEvaluator evaluator
        -PersonaStore personaStore
        -PersonaPromptBuilder promptBuilder
        -PersonaConfig config
        +ProposeJobs(...)
        +SaveResults(...)
        +GetPersona(...)
    }

    class DialogueCollector {
        <<Interface>>
        +CollectByNPC(ctx context.Context, input PersonaGenInput) ([]NPCDialogueData, error)
    }

    class ImportanceScorer {
        <<Interface>>
        +Score(englishText string, questID *string, isServicesBranch bool) int
    }

    class ContextEvaluator {
        <<Interface>>
        +Evaluate(dialogueData NPCDialogueData, config PersonaConfig) (TokenEstimation, []DialogueEntry)
    }

    class PersonaStore {
        <<Interface>>
        +InitSchema(ctx context.Context) error
        +SavePersona(ctx context.Context, result PersonaResult) error
        +GetPersona(ctx context.Context, speakerID string) (string, error)
    }

    class PersonaPromptBuilder {
        +BuildSystemPrompt() string
        +BuildUserPrompt(data NPCDialogueData) string
    }

    class TokenEstimator {
        <<Interface>>
        +Estimate(text string) int
    }

    class TokenEstimation {
        +Int InputTokens
        +Int OutputTokens
        +Int TotalTokens
        +Bool ExceedsLimit
    }

    class Pipeline {
        -NPCPersonaGenerator personaGenerator
        -JobQueue jobQueue
        +HandlePersonaGeneration(w http.ResponseWriter, r *http.Request)
    }

    Pipeline --> NPCPersonaGenerator : uses
    NPCPersonaGenerator <|.. NPCPersonaGeneratorImpl : implements
    NPCPersonaGeneratorImpl --> DialogueCollector : uses
    NPCPersonaGeneratorImpl --> ContextEvaluator : uses
    NPCPersonaGeneratorImpl --> PersonaStore : uses
    NPCPersonaGeneratorImpl --> PersonaPromptBuilder : uses
    DialogueCollector <|.. DialogueCollectorImpl : implements
    DialogueCollectorImpl --> ImportanceScorer : uses
    ContextEvaluator <|.. ContextEvaluatorImpl : implements
    ContextEvaluatorImpl --> TokenEstimator : uses
    PersonaStore <|.. SQLitePersonaStore : implements
```

## アーキテクチャの補足：2フェーズモデル (Propose/Save)
本スライスはバッチAPIや長時間実行ジョブに対応するため、**「プロンプト生成(ProposeJobs)」**と**「結果保存(SaveResults)」**の2フェーズに分割されている。
- **Phase 1 (Propose)**: 入力データを解析し、NPCごとに会話データの収集・選別、コンテキスト長評価を行い、LLMリクエスト群を生成する。
- **Phase 2 (Save)**: JobQueue等を通じて取得されたLLMのレスポンス群を受け取り、パースしてペルソナDBに永続化する。

スライス自身はLLM Clientを直接呼び出さず、通信制御はオーケストレーターに委譲される。
