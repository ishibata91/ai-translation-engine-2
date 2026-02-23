# データローダー シーケンス図

> ファイル読み込み → バリデーション → メモリストア格納
> Interface-First: Parser は Translatable Contract のみに依存

---

## 1. メインフロー

```mermaid
sequenceDiagram
    autonumber
    participant C as Caller
    participant L as parser
    participant FS as FileSystem
    participant J as JSON Decoder

    C ->> L: LoadExtractedJSON(path)
    L ->> FS: Stat(path)

    alt ファイル未検出
        FS -->> L: error
        L -->> C: nil, FileNotFoundError
    end

    L ->> FS: Open(path)
    FS -->> L: io.Reader
```

---

## 2. エンコーディング試行

```mermaid
sequenceDiagram
    autonumber 4
    participant L as parser
    participant J as JSON Decoder

    rect rgb(240, 248, 255)
        Note over L, J: loop: UTF-8 → BOM → CP1252 → Shift-JIS
        L ->> J: NewDecoder(reader)
        L ->> J: Decode(&rawData)
        alt 成功
            J -->> L: rawData
        else 失敗
            J -->> L: error → 次へ
        end
    end

    alt 全て失敗
        L -->> L: return nil, ValueError
    end
```

---

## 3. カテゴリ別パース

```mermaid
sequenceDiagram
    autonumber 7
    participant L as parser
    participant E as Normalizer
    participant P as Parser
    participant S as ExtractedData

    rect rgb(245, 255, 245)
        Note over L, P: par: goroutine 並列化可能

        L ->> P: parseDialogueGroups()
        P ->> E: normalizeTextFields()
        E -->> P: 正規化済み
        P ->> P: Unmarshal + Validate

        alt バリデーション失敗
            Note right of P: Warn → skip
        end

        P -->> L: []DialogueGroup

        L ->> P: parseQuests()
        P -->> L: []Quest

        L ->> P: parse (Items/Magic/...)
        P -->> L: 各 []Struct

        L ->> P: parseNPCs()
        P ->> E: extractEditorID()
        E -->> P: EditorID
        P -->> L: map[string]NPC
    end

    L ->> S: ExtractedData{...}
    S -->> L: *ExtractedData
    L -->> L: return (*ExtractedData, nil)
```

---

## エラーハンドリング方針

| エラー              | タイミング     | 対処        |
| ------------------- | -------------- | ----------- |
| FileNotFound        | ファイル確認   | 即 return   |
| JSONDecode          | エンコード試行 | 次を再試行  |
| Validation (個別)   | パース中       | Warn & skip |
| Validation (ルート) | 構築時         | 即 return   |

---

## Go 移行ポイント

| 項目       | Python → Go                             |
| ---------- | --------------------------------------- |
| JSON       | `json.load()` → `encoding/json` Decoder |
| 検証       | Pydantic → custom Unmarshal             |
| 文字コード | latin-1 fallback → `x/text/encoding`    |
| 並列化     | なし → goroutine + WaitGroup            |
| エラー     | try/except → error + continue           |
