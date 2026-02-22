## ADDED Requirements

### Requirement: PersonaGenSliceの独立した初期化
**Reason**: 垂直スライス (Vertical Slice) の自律性を確保するため、PersonaGenSlice は `ExtractedData` のようなグローバルなデータモデルに依存してはならない。
**Migration**: `PersonaGenInput` DTO とそれを受け取るメソッドを導入する。

#### Scenario: オーケストレーターがマッピングされたデータをPersonaGenSliceに渡す
- **WHEN** オーケストレーターが LoaderSlice の出力から会話を抽出し、`PersonaGenInput` にマッピングして渡す
- **THEN** PersonaGenSlice が正常に初期化され、スコアリングの準備が整う

### Requirement: 会話の重要度スコアリング
**Reason**: LLM のコンテキストウィンドウには制限があるため、ペルソナ生成に最も関連性の高い会話を選択する必要がある。
**Migration**: 固有名詞と感情指標に基づいて会話をスコアリングする新機能。

#### Scenario: 固有名詞と感情による会話のスコアリング
- **WHEN** 会話に 2 つの固有名詞と 1 つの感情指標（例：すべて大文字）が含まれる
- **THEN** 重要度スコアは `2 * W_noun + 1 * W_emotion + base_priority` として計算される

### Requirement: コンテキスト長制限の安全性のためのトークン推定
**Reason**: トークン制限による LLM API エラーを防ぐため。
**Migration**: 新機能。

#### Scenario: 推定トークン数が上限を超える
- **WHEN** 特定の NPC の会話の合計推定トークン数が設定された LLM コンテキストウィンドウを超える
- **THEN** スライスは、合計トークン数が上限内に収まるまで、スコアの低い会話を除外する

#### Scenario: 推定トークン数が上限内に収まる
- **WHEN** 特定の NPC の会話の合計推定トークン数が設定された LLM コンテキストウィンドウ内に収まる
- **THEN** スライスはペルソナ生成のためにすべての会話（最大 100 件）を保持する

### Requirement: LLM ペルソナ生成
**Reason**: 翻訳時の一貫した口調を保つため、各 NPC について簡潔なペルソナテキストを生成する。
**Migration**: LLMManager を使用する新機能。

#### Scenario: ペルソナテキストの生成成功
- **WHEN** 抽出された会話の上位データに基づいて LLM が NPC のペルソナテキストを生成する
- **THEN** テキストは特定のフォーマット（性格特性, 話し方の癖, 背景設定）に従っていること

### Requirement: ペルソナの永続化 (PersonaStore)
**Reason**: 冗長な LLM リクエストを避けるため、生成されたペルソナをキャッシュする。
**Migration**: SQLite の `npc_personas` テーブルを使用し、PersonaGenSlice 内にカプセル化された新機能。

#### Scenario: 生成されたペルソナの Upsert
- **WHEN** `speaker_id` に対してペルソナの生成が成功する
- **THEN** PersonaStore は提供された `*sql.DB` 接続を使用して、`npc_personas` テーブルにレコードを Upsert する

### Requirement: 構造化ログの適用
**Reason**: ユニットテストが除外されているため、標準的なデバッグワークフローに従う必要がある。
**Migration**: `slog.DebugContext` を使用した Entry/Exit ロギングを実装する。

#### Scenario: 生成開始時のログ出力
- **WHEN** 処理が開始される
- **THEN** DEBUG レベルで入力の会話件数と `trace_id` を含んだ Entry の JSON ログが出力される
