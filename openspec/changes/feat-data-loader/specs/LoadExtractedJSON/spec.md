# LoadExtractedJSON Spec

## ADDED Requirements

### Requirement: JSON Loading

抽出されたJSONファイルを読み込み、`ExtractedData` 構造体にマッピングする。

#### Scenario: Successful Load
- **WHEN** 有効なパスと有効なJSONファイルが渡されたとき
- **THEN** エラーなし (`nil`) を返す
- **AND** 返却される `*ExtractedData` 構造体の各フィールド (Quests, NPC, Items 等) にデータが格納されている

#### Scenario: File Not Found
- **WHEN** 存在しないファイルパスが渡されたとき
- **THEN** `os.ErrNotExist` ラップしたエラーを返す
- **AND** `nil` を返す

#### Scenario: Invalid JSON Syntax
- **WHEN** JSON構文が壊れているファイルが渡されたとき
- **THEN** JSONデコードエラーを返す
- **AND** `nil` を返す

#### Scenario: Partial Data Loading
- **WHEN** `quests` フィールドのみが存在するJSONが渡されたとき
- **THEN** エラーなしを返す
- **AND** `ExtractedData.Quests` にデータが格納されている
- **AND** 他のフィールド (NPCs, Items 等) は空またはゼロ値である
