# PersonaGenSlice 仕様書

## Purpose
PersonaGenSlice は、NPC（Non-Player Character）の会話履歴を分析し、それぞれの性格や話し方の特徴（ペルソナ）を要約して生成・保存する責務を持ちます。生成されたペルソナデータは、後続の翻訳プロセスにおいて、キャラクターの口調や一貫性を保つための重要なコンテキストとして利用されます。

## Requirements

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

### Requirement: LLM ペルソナ生成 (Phase 1)
**Reason**: 翻訳時の一貫した口調を保つため、各 NPC について簡潔なペルソナテキストを生成するためのプロンプトを構築する。
**Migration**: 自身でLLMを呼び出すのではなく、JobQueueに渡すための `[]llm_client.Request` を返す純粋関数として再設計。

#### Scenario: 抽出データからのリクエスト構築
- **WHEN** プロセスマネージャーから `PersonaGenInput` を受け取った
- **THEN** 内部のDBを参照し、すでに作成済みのペルソナを除外する
- **AND** 会話の重要度スコアとトークン数を計算し、条件に合致するものだけのプロンプトを構築する
- **AND** 作成したプロンプト配列 `[]llm_client.Request` を返す

### Requirement: ペルソナ結果の抽出と保存 (Phase 2)
**Reason**: JobQueueから返されたLLMレスポンスをパースし、安全に永続化する。
**Migration**: 出力フォーマットを厳格化し、堅牢な抽出ロジックと個別エラーハンドリングを備えた保存フェーズを導入。

#### 抽出フォーマットの統一
全てのプロンプトにおいて、出力フォーマットを `TL: |ペルソナ内容|` のパイプ区切り形式に強制する。システムは正規表現（`TL:\s*\|(.*?)\|`）を優先し、複数のフォールバック（プレフィックス除去、全体トリミング）を用いてテキストを抽出しなければならない。

#### Scenario: 成功レスポンスのパースと保存
- **WHEN** プロセスマネージャーから `[]llm_client.Response` が渡された
- **THEN** 各レスポンスの `Metadata` から対象NPCを特定し、規定の形式に従ってペルソナ文を抽出する
- **AND** 抽出に成功したペルソナを、`npc_personas` テーブルに UPSERT する
- **AND** 個別のパースエラーやDBエラーが発生しても処理を中断せず、可能な限り多くのデータを保存し、警告ログを出力する

### Requirement: 構造化ログの適用
**Reason**: ユニットテストが除外されているため、標準的なデバッグワークフローに従う必要がある。
**Migration**: `slog.DebugContext` を使用した Entry/Exit ロギングを実装する。

#### Scenario: 生成開始時のログ出力
- **WHEN** 処理が開始される
- **THEN** DEBUG レベルで入力の会話件数と `trace_id` を含んだ Entry の JSON ログが出力される
