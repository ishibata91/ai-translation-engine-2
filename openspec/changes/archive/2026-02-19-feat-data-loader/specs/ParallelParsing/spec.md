# ParallelParsing Spec

## ADDED Requirements

### Requirement: Parallel Mapping & Normalization

JSONのロード自体はシングルスレッドで行い、その後の構造体へのマッピング（Unmarshal）と正規化処理を並列化する。

#### Scenario: Two-Phase Loading
- **WHEN** `LoadExtractedJSON` が呼び出されたとき
- **THEN** まず `map[string]json.RawMessage` としてトップレベルのキーをデコードする (Serial)
- **AND** その後、`quests`, `dialogue_groups` などの重いフィールドの Unmarshal と正規化を個別の Goroutine で開始する (Parallel)

#### Scenario: Normalization Overhead
- **WHEN** 正規化処理（ID抽出やバリデーション）が走るとき
- **THEN** 各Goroutine内で並列に実行され、メインスレッドのブロックを防ぐ

#### Scenario: Error Aggregation
- **WHEN** 並列処理中に1つ以上のマッピングエラーが発生したとき
- **THEN** エラーを収集し、処理完了後にまとめて（または最初のクリティカルエラーを）返す
