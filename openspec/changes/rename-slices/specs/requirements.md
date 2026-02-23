## MODIFIED Requirements

### 2.2 アーキテクチャ: 2-Pass System
品質と一貫性を担保するため、処理を2段階に分割する。

*   **Pass 1: 用語翻訳と構造化 (Term Translation & Structuring)**
    *   固有名詞（NPC名、アイテム名、場所名）を先行して翻訳し、辞書キャッシュを構築する（`dictionary`, `terminology`）。
    *   会話データ等の依存関係を解析し、構造化データを作成する（`parser`）。
*   **Pass 2: 本文翻訳 (Main Translation)**
    *   Pass 1で構築した用語辞書（`dictionary`, `terminology`）と文脈情報（`lore`, `persona`, `summary`）を用いて、会話文や説明文を翻訳する（`translator`）。

#### Scenario: 2-Passの実行
- **WHEN** 全体プロセスが開始される
- **THEN** Pass 1 (`parser`, `dictionary`, `terminology`) が実行され、その後に Pass 2 (`translator`) が実行される。

### 3.1 データ入力とロード
*   **抽出データ読み込み**: xEditスクリプトによって出力されたデータ（JSON等）を読み込むことができること（`parser`）。
*   **複数ファイル対応**: 複数の入力ファイルを一括でロードし、処理できること。
*   **データバリデーション**: 読み込み時にデータの整合性をチェックし、不正なレコードを除外または報告すること。
*   **外部辞書作成 (Dictionary)**: xTranslator形式のXMLファイルを読み込み、名詞類（NPC名、アイテム名など）を抽出してSQLiteベースの用語辞書DBを構築する機能。

#### Scenario: XMLデータのロード
- **WHEN** xTranslator形式のXMLが指定される
- **THEN** `dictionary` スライスが内容を解析し、DBに保存する。

### 3.2 コンテキスト構築 (Context Building)
*   **会話ツリー解析**: `DIAL/INFO` レコードの親子関係 (`PNAM`) を追跡し、直前の発言（Previous Line）を特定できること（`lore`）。
*   **話者プロファイリング (Speaker Profiling)**:
    *   NPCの種族 (`RNAM`)、声の種類 (`VTCK`)、性別フラグ (`ACBS`) から、適切な口調（Tone）を推定すること。
    *   例：「カジート」→「〜とお前は思う」、「オーク」→「粗野な口調」。
    *   ユーザーによるカスタマイズ設定（特定のNPCへの口調強制など）を可能にすること。
*   **NPCペルソナ生成 (Persona)**:
    *   NPCごとに会話データ（`DIAL/INFO`）を最大100件まで収集し、LLMにリクエストしてそのNPCの性格・口調・背景を要約した「ペルソナ」を自動生成すること（`persona`）。
*   **クエスト要約 (Summary)**:
    *   クエストステージ (`QUST INDX`) を時系列順に処理し、過去のステージの翻訳結果を「これまでのあらすじ」として累積的に次ステージのコンテキストに含めること（`summary`）。

#### Scenario: ペルソナの活用
- **WHEN** 通過2の翻訳が実行される
- **THEN** `persona` スライスによって生成されたペルソナがプロンプトに組み込まれる。

### 3.8 用語適用 (Terminology)
*   **用語の適用**: 翻訳対象テキストに対して、辞書から検索された適切な訳語を適用すること。

#### Scenario: 用語の検索
- **WHEN** 翻訳対象テキストが `terminology` スライスに渡される
- **THEN** 最適な用語ペアが返される。

### その他メンテナンス
- 共通ドキュメント `specs/refactoring_strategy.md` は `specs/architecture.md` にリネームされ、システム全体のアーキテクチャ定義として更新される。
