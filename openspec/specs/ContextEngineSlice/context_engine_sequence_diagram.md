# コンテキストエンジン シーケンス図

## 1. 全体フロー（BuildTranslationRequests）

```mermaid
sequenceDiagram
    participant PM as ProcessManager
    participant CE as ContextEngineImpl
    participant TR as ToneResolver
    participant PL as PersonaLookup
    participant TL as TermLookup
    participant SL as SummaryLookup

    PM->>CE: BuildTranslationRequests(data, config)
    CE->>CE: buildDialogueRequests(groups, npcs, config)
    CE->>CE: buildQuestRequests(quests, config)
    CE->>CE: buildItemRequests(items, config)
    CE->>CE: buildMagicRequests(magics, config)
    CE->>CE: buildMessageRequests(messages, config)
    CE-->>PM: []TranslationRequest
```

## 2. 会話リクエスト構築フロー（buildDialogueRequests）

```mermaid
sequenceDiagram
    participant CE as ContextEngineImpl
    participant SL as SummaryLookup
    participant TR as ToneResolver
    participant PL as PersonaLookup
    participant TL as TermLookup

    loop for each DialogueGroup
        CE->>SL: FindDialogueSummary(groupID)
        SL-->>CE: *string / nil

        Note over CE: last_line = group.PlayerText

        opt PlayerText exists: DIAL FULL request
            CE->>TL: Search(playerText)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest (DIAL FULL)
        end

        loop for each DialogueResponse (Order昇順)
            CE->>CE: resolveSpeaker(speakerID, npcs)
            CE->>TR: Resolve(race, voiceType, sex)
            TR-->>CE: toneInstruction
            CE->>PL: FindBySpeakerID(speakerID)
            PL-->>CE: *personaText / nil
            Note over CE: → SpeakerProfile
            CE->>CE: getTopicName(response, group)

            opt INFO RNAM: prompt/選択肢
                CE->>TL: Search(promptText)
                TL-->>CE: []ReferenceTerm, *forcedTranslation
                Note over CE: → TranslationRequest (INFO RNAM, previousLine=last_line)
                Note over CE: last_line = menuDisplayText or prompt
            end

            opt INFO NAM1: NPCセリフ
                CE->>TL: Search(responseText)
                TL-->>CE: []ReferenceTerm, *forcedTranslation
                Note over CE: → TranslationRequest (INFO NAM1, speaker, previousLine=last_line)
                Note over CE: last_line = response.Text
            end
        end
    end
```

## 3. クエストリクエスト構築フロー（buildQuestRequests）

```mermaid
sequenceDiagram
    participant CE as ContextEngineImpl
    participant SL as SummaryLookup
    participant TL as TermLookup

    loop for each Quest
        CE->>SL: FindQuestSummary(questID)
        SL-->>CE: *string / nil

        opt Quest.Name exists: QUST FULL
            CE->>TL: Search(questName)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest (QUST FULL)
        end

        loop for each Stage (Index昇順)
            CE->>TL: Search(stageText)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest (QUST CNAM, questSummary, index)
        end

        loop for each Objective
            CE->>TL: Search(objectiveText)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest (QUST NNAM, questSummary, index)
        end
    end
```

## 4. アイテム・書籍リクエスト構築フロー（buildItemRequests）

```mermaid
sequenceDiagram
    participant CE as ContextEngineImpl
    participant TL as TermLookup

    loop for each Item
        opt Item.Description exists: {Type} DESC
            CE->>TL: Search(description)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest ({Type} DESC)
        end

        opt Item.Text exists (BOOK): BOOK DESC
            CE->>TL: Search(bookText)
            TL-->>CE: []ReferenceTerm, *forcedTranslation
            Note over CE: → TranslationRequest (BOOK DESC, maxTokens)
        end
    end
```

## 5. 話者解決フロー（resolveSpeaker）

```mermaid
sequenceDiagram
    participant CE as ContextEngineImpl
    participant TR as ToneResolver
    participant PL as PersonaLookup

    Note over CE: resolveSpeaker(speakerID, npcs)

    alt speakerID == nil
        CE-->>CE: return nil
    else NPC found in npcs map
        CE->>TR: Resolve(npc.Race, npc.Voice, npc.Sex)
        TR-->>CE: toneInstruction
        CE->>PL: FindBySpeakerID(speakerID)
        PL-->>CE: *personaText / nil
        Note over CE: → SpeakerProfile {Name, Gender, Race, VoiceType, ToneInstruction, PersonaText}
    else NPC not found
        CE-->>CE: return nil
    end
```
