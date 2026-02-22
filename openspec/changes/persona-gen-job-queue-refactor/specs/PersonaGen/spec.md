# PersonaGen Delta Spec

## MODIFIED Requirements

### Requirement: プロンプト生成フェーズ (Phase 1)
PersonaGenSliceは、渡された入力からプロンプトを構築し、インフラに引き渡すための `[]llm_client.Request` を返す純関数として機能しなければならない。既存の「作成済みスキップ」や「トークン数に基づく不要なプロンプトの除外」ロジックは維持する。

#### Scenario: 抽出データからのリクエスト構築
- **WHEN** プロセスマネージャーから `PersonaGenInput` を受け取った
- **THEN** 内部のDB（`Mod用語DB`）を参照し、すでに作成済みのペルソナを除外する
- **AND** セリフのトークン数を計算し、条件に合致するものだけのプロンプトを構築する
- **AND** 作成したプロンプト配列 `[]llm_client.Request` を返す

### Requirement: ペルソナ結果の保存フェーズ (Phase 2)
PersonaGenSliceは、JobQueueからコールバックされた `[]llm_client.Response` をパースし、JSON形式等からペルソナ情報を抽出して `Mod用語DB` に安全にUPSERTしなければならない。

#### Scenario: 成功レスポンスのパースと保存
- **WHEN** プロセスマネージャーから `[]llm_client.Response` が渡された
- **THEN** 各レスポンスからフォーマット通りにペルソナ文を抽出する
- **AND** 抽出に成功したペルソナを、対象NPCのレコードとして UPSERT する
- **AND** （失敗した応答やパースエラーはスキップし、成功分のみを保存して処理完了とする）
