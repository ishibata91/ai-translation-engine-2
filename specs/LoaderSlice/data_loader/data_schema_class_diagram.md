# データスキーマ クラス図

> Phase 1: Data Foundation
> Interface-First 設計に基づく Go Struct 移行
> ※ DSDRecord / Translatable は廃止し、ExtractedData を直接利用する

---

## 1. 共通埋め込み構造体

```mermaid
classDiagram
    direction LR

    class BaseExtractedRecord {
        <<embedded>>
        +ID         string
        +EditorID   *string
        +Type       string
        +SourceJSON string
    }
```

全ドメインモデルが上記を埋め込む。

---

## 2. ドメインモデル — 会話系

```mermaid
classDiagram
    direction TB

    class DialogueResponse {
        BaseExtractedRecord
        +Text            string
        +Prompt          *string
        +TopicText       *string
        +MenuDisplayText *string
        +SpeakerID       *string
        +VoiceType       *string
        +Order           int
        +PreviousID      *string
        +Source           *string
        +Index           *int
    }

    class DialogueGroup {
        BaseExtractedRecord
        +PlayerText       *string
        +QuestID          *string
        +IsServicesBranch bool
        +ServicesType     *string
        +NAM1             *string
        +Responses        []DialogueResponse
        +Source           *string
    }

    DialogueGroup *-- "0..*" DialogueResponse
```

---

## 3. ドメインモデル — クエスト系

```mermaid
classDiagram
    direction TB

    class QuestStage {
        +Index  int
        +Type   string
        +Text   string
        +Source  *string
    }

    class QuestObjective {
        +Index  int
        +Type   string
        +Text   string
        +Source  *string
    }

    class Quest {
        BaseExtractedRecord
        +Name         *string
        +Stages       []QuestStage
        +Objectives   []QuestObjective
        +Source        *string
    }

    Quest *-- "0..*" QuestStage
    Quest *-- "0..*" QuestObjective
```

---

## 4. ドメインモデル — エンティティ系

```mermaid
classDiagram
    direction TB

    class NPC {
        BaseExtractedRecord
        +Name       string
        +Race       string
        +Voice      string
        +Sex        string
        +ClassName   *string
        +Source      *string
        +IsFemale()  bool
    }

    class Item {
        <<WEAP ARMO BOOK MISC>>
        BaseExtractedRecord
        +Name         *string
        +Description  *string
        +Text         *string
        +TypeHint     *string
        +Source       *string
    }

    class Magic {
        <<SPEL MGEF ENCH SHOU>>
        BaseExtractedRecord
        +Name         *string
        +Description  *string
        +Source       *string
    }
```

---

## 5. ドメインモデル — その他

```mermaid
classDiagram
    direction TB

    class Location {
        <<LCTN WRLD CELL>>
        BaseExtractedRecord
        +Name      *string
        +ParentID  *string
        +Source    *string
    }

    class Message {
        <<MESG>>
        BaseExtractedRecord
        +Text     string
        +Title    *string
        +QuestID  *string
        +Source   *string
    }

    class SystemRecord {
        <<PERK etc>>
        BaseExtractedRecord
        +Name         *string
        +Description  *string
        +Source       *string
    }

    class LoadScreen {
        <<LSCR>>
        BaseExtractedRecord
        +Text    string
        +Source  *string
    }
```

---

## 6. ルートコンテナ & 依存関係

```mermaid
classDiagram
    direction TB

    class ExtractedData {
        <<Root Container>>
        +DialogueGroups []DialogueGroup
        +Quests         []Quest
        +Items          []Item
        +Magic          []Magic
        +Locations      []Location
        +Cells          []Location
        +System         []SystemRecord
        +Messages       []Message
        +LoadScreens    []LoadScreen
        +NPCs           map_string_NPC
    }
```

---

## 凡例

| 記号           | 意味                  |
| -------------- | --------------------- |
| `*--`          | コンポジション (所有) |
| `<<embedded>>` | Go struct 埋め込み    |
