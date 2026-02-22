# コンテキストエンジン (Context Engine Slice) 仕様書

## 概要
翻訳対象レコードに対して、AIが高品質な翻訳を行うために必要な「文脈情報（コンテキスト）」を構築する機能である。
会話ツリーの解析、話者プロファイリング、参照用語の検索、および各レコードタイプに応じた翻訳リクエストの組み立てを担い、Pass 2（本文翻訳）Sliceに対して構造化された `TranslationRequest` のリストを提供する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「翻訳リクエストの構築ロジック」「会話ツリー解析」「話者プロファイリング」「参照用語検索」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `context_builder.py` が2,300行超の巨大モジュールとして、用語翻訳リクエスト生成（Pass 1）と本文翻訳リクエスト生成（Pass 2）の両方を担っている。
- Go v2では責務を明確に分離し、用語翻訳リクエスト生成は Term Translator Slice に、本文翻訳リクエスト生成は本 Context Engine Slice に分割する。
- 要件定義書 §3.2「コンテキスト構築」に基づき、会話ツリー解析・話者プロファイリング・クエスト要約参照・用語検索を統合した構造化コンテキストを構築する。

## スコープ
### 本Sliceが担う責務
1. **会話ツリー解析**: `DialogueGroup` / `DialogueResponse` の `PreviousID` チェーンをトラバースし、直前の発言（Previous Line）を特定する。
2. **話者プロファイリング**: NPCの種族・性別・声タイプからの口調推定、およびペルソナ（Persona Generator Slice生成）の参照。
3. **参照用語検索**: 翻訳対象テキストからキーワードを抽出し、辞書DB（Dictionary Builder Slice構築）およびMod用語DB（Term Translator Slice構築）から関連用語を検索する。
4. **翻訳リクエストの構築**: 上記のコンテキスト情報を統合し、レコードタイプ別の `TranslationRequest` を生成する。
5. **要約の参照**: Summary Generator Sliceが生成した会話要約・クエスト要約をSQLiteから取得し、コンテキストに含める。

### 本Sliceの責務外
- 用語翻訳（Term Translator Sliceの責務）
- NPCペルソナ生成（Persona Generator Sliceの責務）
- 会話・クエスト要約の生成（Summary Generator Sliceの責務）
- LLMによる本文翻訳の実行（Pass 2 Sliceの責務）
- xTranslator XMLからの辞書DB構築（Dictionary Builder Sliceの責務）

## 要件

### 独立性: コンテキストエンジン向けデータの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライス(LoaderSlice等)のDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTOを定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化とコンテキスト構築
- **WHEN** オーケストレーター層から本スライス専用の入力DTOが提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いて文脈解析・構築処理を完結できること

### 1. 翻訳リクエストのデータ構造

本Sliceが生成する翻訳リクエストの構造を以下に定義する。

**`TranslationRequest` 構造体**:
```go
// TranslationRequest はPass 2翻訳の単位となるリクエストを表す。
type TranslationRequest struct {
    ID                string              // FormID
    RecordType        string              // レコードタイプ（例: "INFO NAM1", "QUST CNAM"）
    SourceText        string              // 翻訳対象テキスト（英語原文）
    Context           TranslationContext  // 翻訳コンテキスト
    Index             *int                // Stage/Objective Index（クエスト用）
    ReferenceTerms    []ReferenceTerm     // 参照用語リスト
    EditorID          *string             // Editor ID
    ForcedTranslation *string             // 完全一致時の強制翻訳結果
    SourcePlugin      string              // ソースプラグイン名
    SourceFile        string              // リクエスト発生元のファイル名
    MaxTokens         *int                // 最大トークン数（動的計算結果）
}
```

**`TranslationContext` 構造体**:
```go
// TranslationContext は翻訳に必要な文脈情報を保持する。
type TranslationContext struct {
    PreviousLine    *string         // 会話: 直前のセリフ
    Speaker         *SpeakerProfile // 会話: 話者属性
    TopicName       *string         // 会話: トピック名
    QuestName       *string         // クエスト: クエスト名
    QuestSummary    *string         // クエスト: クエスト全体の要約
    DialogueSummary *string         // 会話: 会話全体の要約
    ItemTypeHint    *string         // アイテム: 種別ヒント
    ModDescription  *string         // Mod全体の説明
    PlayerTone      *string         // プレイヤーの口調指示
}
```

**`SpeakerProfile` 構造体**:
```go
// SpeakerProfile はNPCの話者属性を表す。
type SpeakerProfile struct {
    Name             string  // NPC名
    Gender           string  // 性別（"Male" / "Female"）
    Race             string  // 種族名
    VoiceType        string  // ボイスタイプ
    ToneInstruction  string  // 口調指示文
    PersonaText      *string // ペルソナテキスト（Persona Generator Slice生成、存在する場合）
}
```

**`ReferenceTerm` 構造体**:
```go
// ReferenceTerm は参照用語（辞書からの既訳）を表す。
type ReferenceTerm struct {
    OriginalEN string // 英語原文
    OriginalJA string // 日本語訳
}
```

### 2. 会話ツリー解析

`DialogueGroup` および `DialogueResponse` の情報を基に会話ツリーを有向非巡回グラフ (DAG) として捉え、`Previous Lines` トラバースによる直前の発言を特定する。

**トラバースアルゴリズムの設計**:
1. **DAGベースの探索**: `DialogueResponse.PreviousID` のチェーンをたどり、各会話ノードのリンクを解析する。
2. **分岐と合流の処理**:
    - 複数のレスポンスが同じグループ（発言）を指す場合など、複雑な会話パスにおいては、`Order` 昇順で最後に解決されたパス、またはデフォルトルートルート（ゲーム内で主要とされる進行ルート）の `PreviousLine` を採用する。
3. **循環参照（無限ループ）の防止**: トラバース中に無限ループを検知・回避するため、探索の深さ制限（Max Depth: Config定義、デフォルト5）を設ける。制限に達した場合はトラバースを打ち切り、そこまでのルートを採用する。

**Previous Line の決定ルール**:
1. 各 `DialogueGroup` の処理開始時、`last_line` を `DialogueGroup.PlayerText`（プレイヤーの発言）で初期化する。`PlayerText` が nil の場合は空文字列とする。
2. 上記のトラバースアルゴリズムを加味し、`DialogueResponse` を順次処理する。
3. 各 `DialogueResponse` の `INFO RNAM`（プロンプト/選択肢）リクエスト生成後、`last_line` を当該レスポンスの `MenuDisplayText` または `Prompt` で更新する。
4. 各 `DialogueResponse` の `INFO NAM1`（NPCセリフ）リクエスト生成後、`last_line` を当該レスポンスの `Text` で更新する。

**トピック名の決定ルール**:
1. `DialogueResponse.TopicText` を最優先で使用する。
2. 存在しない場合、`DialogueGroup.NAM1` を使用する。
3. それも存在しない場合、`DialogueGroup.EditorID` を使用する。
4. いずれも存在しない場合、`"（名称未設定）"` をフォールバックとする。
5. トピック名が100文字を超える場合は97文字に切り詰めて `"..."` を付加する。

### 3. 話者プロファイリング

NPCの属性（種族・クラス等）と性格データ（ペルソナ）を統合して最適な口調指示文を生成し、`SpeakerProfile` を構築する。

**口調モデリングアルゴリズム（優先度順）**:
1. **ペルソナ（性格データ）によるオーバーライド**: 
   - Persona Generator Slice が生成した `npc_personas` テーブルを `SpeakerID` で検索し、固有の性格・背景テキストが存在する場合、これが最も強い口調の文脈として適用される（種族やクラスの一般的な口調を上書き、または補強する）。
2. **種族ベースのマッピング**: NPCの種族（`Race`）に基づく口調指示（Config定義）。
   - 例: カジート → 独特の三人称話法、オーク → 粗野な口調、アルトマー → 高慢な口調。
3. **クラス（職業）ベースのマッピング**: NPCのクラス（`Class` / 戦士、魔術師、貴族など）に基づく口調の微調整（Config定義）。
   - 例: 種族がノルドでも、クラスが貴族であれば、野蛮な口調を抑え知的な言い回しを含める。
4. **ボイスタイプベースのマッピング**: NPCのボイスタイプ（`Voice`）に基づく口調指示。
5. **性別ベースのフォールバック**: 上記情報が不十分な場合、性別に基づく標準口調を適用。
   - Female: `"標準的な女性の口調（一人称は主に「私」を使い、落ち着いた知的な、あるいは世話好きな態度）"`
   - Male: `"標準的な男性の口調（一人称は状況に応じて「私」または「俺」を使い、粗暴すぎず丁寧すぎない、落ち着いた態度）"`

**プロファイルの合成**: 
ペルソナテキストが存在する場合はそれを主体とし、存在しない場合は種族・クラス・ボイスの各指示を組み合わせて、LLMが理解しやすいプレーンな「口調指示文」として `ToneInstruction` に結合する。

**`ToneResolver` インターフェース**:
```go
// ToneResolver はNPC属性から口調指示文を生成する。
type ToneResolver interface {
    Resolve(race string, class string, voiceType string, sex string) string
}
```

**`PersonaLookup` インターフェース**:
```go
// PersonaLookup はNPCのペルソナテキストを検索する。
type PersonaLookup interface {
    FindBySpeakerID(speakerID string) (*string, error)
}
```

### 4. 参照用語検索

翻訳対象テキストからキーワードを抽出し、辞書DB（Dictionary Builder Slice構築）およびMod用語DB（Term Translator Slice構築）から関連用語を検索する。

**検索戦略**:
本Sliceの辞書検索は、Term Translator Sliceの検索戦略（§2 貪欲部分一致）と同一のアルゴリズムを採用する。ただし、Vertical Sliceの自律性原則に従い、検索ロジックは本Slice内に独立して実装する（WET原則）。

**検索フロー（優先順）**:
1. **ソーステキスト全文の完全一致**: 翻訳対象テキスト全体を辞書の `source` カラムと照合する。
2. **キーワード完全一致**: ソーステキストから抽出したキーワード群を辞書の `source` カラムと `IN` 句で照合する。
3. **NPC名の貪欲部分一致**: NPC名専用FTS5テーブルに対する部分一致検索。
4. **貪欲最長一致フィルタリング**: 候補を文字数降順でソートし、重複する文字区間を排除する。

**ステミング**: Term Translator Sliceと同様、Snowball English Stemmerを適用してキーワードの形態変化に対応する。

**強制翻訳（Forced Translation）**:
- 辞書DBにソーステキスト全文の完全一致する既訳が存在する場合、`TranslationRequest.ForcedTranslation` に辞書の訳を設定する。Pass 2 SliceはこのフィールドがnilでなければLLM呼び出しをスキップし、辞書の訳をそのまま採用する。

**`TermLookup` インターフェース**:
```go
// TermLookup は翻訳対象テキストに関連する参照用語を検索する。
type TermLookup interface {
    Search(sourceText string) ([]ReferenceTerm, *string, error)
    // Returns: (参照用語リスト, 強制翻訳結果（完全一致時）, エラー)
}
```

**DB接続**:
- 辞書DB（`*sql.DB`）およびMod用語DB（`*sql.DB`）をDIで受け取る。
- 追加辞書DB（`additional_db_paths`）にも対応し、複数DBを横断検索する。

### 5. レコードタイプ別の翻訳リクエスト構築

本Sliceは、`ExtractedData` に含まれる各ドメインモデルからレコードタイプに応じた `TranslationRequest` を生成する。

#### 5.1 会話レコード（DIAL/INFO）

| レコードタイプ | ソース                                         | 説明                                         |
| :------------- | :--------------------------------------------- | :------------------------------------------- |
| `DIAL FULL`    | `DialogueGroup.PlayerText`                     | プレイヤーの選択肢テキスト（グループレベル） |
| `INFO RNAM`    | `DialogueResponse.MenuDisplayText` or `Prompt` | プレイヤーの選択肢/プロンプト                |
| `INFO NAM1`    | `DialogueResponse.Text`                        | NPCのセリフ                                  |

**コンテキスト構築**:
- `PreviousLine`: §2 の会話ツリー解析で決定。
- `Speaker`: `INFO NAM1` の場合、`DialogueResponse.SpeakerID` から NPC を解決し、§3 の話者プロファイリングで `SpeakerProfile` を構築する。`DIAL FULL` / `INFO RNAM` の場合は nil（プレイヤー発言）。
- `TopicName`: §2 のトピック名決定ルールに従う。
- `DialogueSummary`: Summary Generator Sliceが生成した会話要約を参照する。
- `ModDescription`: Config定義のMod説明文。
- `PlayerTone`: Config定義のプレイヤー口調指示。

#### 5.2 クエストレコード（QUST）

| レコードタイプ | ソース                | 説明                 |
| :------------- | :-------------------- | :------------------- |
| `QUST FULL`    | `Quest.Name`          | クエスト名           |
| `QUST CNAM`    | `QuestStage.Text`     | クエストステージ記述 |
| `QUST NNAM`    | `QuestObjective.Text` | クエスト目標テキスト |

**コンテキスト構築**:
- `QuestName`: `Quest.Name` を設定する。
- `QuestSummary`: Summary Generator Sliceが生成したクエスト要約を参照する。`QUST CNAM` / `QUST NNAM` の場合に設定し、`QUST FULL` の場合は nil。
- `ModDescription`: Config定義のMod説明文。

#### 5.3 アイテムレコード

| レコードタイプ | ソース             | 説明           |
| :------------- | :----------------- | :------------- |
| `BOOK DESC`    | `Item.Text`        | 書籍の本文     |
| `{Type} DESC`  | `Item.Description` | アイテム説明文 |

**コンテキスト構築**:
- `ItemTypeHint`: `Item.TypeHint` を設定する（例: `"Weapon"`, `"Armor"`, `"Book"`）。
- `ModDescription`: Config定義のMod説明文。

**書籍の長文分割（Chunking）**:
- `BOOK DESC` のテキストが長文の場合、HTMLタグ構造を維持しつつチャンク分割する。
- 分割閾値はConfig定義とする（デフォルト: 1500文字）。
- 各チャンクに対して個別の `TranslationRequest` を生成し、`MaxTokens` を動的に計算する。

#### 5.4 魔法レコード

| レコードタイプ | ソース              | 説明             |
| :------------- | :------------------ | :--------------- |
| `{Type} DESC`  | `Magic.Description` | 魔法効果の説明文 |

#### 5.5 メッセージレコード

| レコードタイプ | ソース                | 説明               |
| :------------- | :-------------------- | :----------------- |
| `MESG DESC`    | `Message.Description` | メッセージの説明文 |

#### 5.6 共通ルール
- 翻訳対象テキストが既に日本語を含む場合はスキップする（`contains_japanese` 判定）。
- 翻訳対象テキストが空または空白のみの場合はスキップする。
- 各リクエストに `SourcePlugin`（ソースプラグイン名）と `SourceFile`（ファイル名）を設定する。

### 6. 要約の参照

Summary Generator Sliceが生成した要約をSQLiteから取得し、コンテキストに含める。

**`SummaryLookup` インターフェース**:
```go
// SummaryLookup は要約テキストを検索する。
type SummaryLookup interface {
    FindDialogueSummary(dialogueGroupID string) (*string, error)
    FindQuestSummary(questID string) (*string, error)
}
```

**参照ルール**:
1. **会話翻訳時**: `DialogueGroup.ID` をキーとして `summary_type = 'dialogue'` の要約を検索する。
2. **クエスト関連翻訳時**: `Quest.ID` をキーとして `summary_type = 'quest'` の要約を検索する。
3. 要約が存在しない場合は nil とし、要約なしでリクエストを構築する。

**DB接続**:
- Summary Generator Sliceが管理するソースファイル単位のSQLiteキャッシュファイルへの `*sql.DB` をDIで受け取る。
- 本Slice内の `SummaryLookup` 実装が SELECT 操作のみを行う（読み取り専用）。

### 7. プロンプトエンジニアリング設計 (Prompt Engineering Template Design)

レコード種別（会話、クエスト、アイテム等）に応じ、`ConfigStore` のテンプレートを呼び出して動的にパラメータを埋め込むことで、LLMへ送信するプロンプトを構築する。

**基本方針**:
- **役割の分離**: 「システムプロンプト」は翻訳者としての役割、前提条件（ペルソナや要約）、出力ルールを定義する。「ユーザープロンプト」は翻訳対象の入力テキストのみを持たせる。
- **変数の埋め込み**: `{VaraibleName}` 形式のプレースホルダーを、本Sliceで収集したコンテキスト情報に置換する。

**システムプロンプトの動的変数定義**:
- `{SpeakerTone}`: 話者の口調、ペルソナ、背景設定のテキスト
- `{PreviousLine}`: 直前のセリフのテキスト
- `{DialogueSummary}`: 会話フローの要約
- `{QuestSummary}`: クエスト全体のこれまでのあらすじ
- `{ItemTypeHint}`: アイテム種別ヒント 
- `{ReferenceTerms}`: 抽出された用語の [英語: 日本語] ペアのリスト

**レコード種別別標準テンプレート例**:
1. **会話系 (INFO / DIAL)**: 会話文のトーンを維持し、直前の流れを汲み取らせる形。
   ```
   You are an expert translator for the game Skyrim. Translate the given text into natural conversational Japanese.
   
   Speaker Persona: {SpeakerTone}
   Previous Line: {PreviousLine}
   Conversation Context: {DialogueSummary}
   
   Rules:
   - Use the designated persona for the speaker's tone.
   - Match the provided terminology exactly: {ReferenceTerms}
   ```
2. **クエスト系 (QUST)**: 要約との整合性を持たせた、ジャーナル風の端的な表現。
   ```
   Translate the quest log/objective into Japanese, keeping it concise and informative as a player's journal.
   
   Quest Context: {QuestSummary}
   Terminology to strictly follow: {ReferenceTerms}
   ```
3. **アイテム・書籍系 (DESC)**: アイテムとしてのフレーバー維持や、書籍らしい文体（段落維持）。
   ```
   Translate the description into Japanese. It is a description for an in-game item.
   Item Type: {ItemTypeHint}
   Terminology to strictly follow: {ReferenceTerms}
   ```

### 8. メインインターフェース

**`ContextEngine` インターフェース**:
```go
// ContextEngine はExtractedDataから翻訳リクエストのリストを構築する。
type ContextEngine interface {
    BuildTranslationRequests(
        data *ExtractedData,
        config ContextEngineConfig,
    ) ([]TranslationRequest, error)
}
```

**`ContextEngineConfig` 構造体**:
```go
// ContextEngineConfig はContext Engineの実行設定を保持する。
type ContextEngineConfig struct {
    ModDescription string   // Mod全体の説明文
    PlayerTone     string   // プレイヤーの口調指示
    SourceFile     string   // ソースファイル名
}
```

### 9. 進捗通知
- 翻訳リクエスト構築の進捗（レコードタイプ別の完了数/総数）をコールバックまたはチャネル経由でProcess Managerに通知し、UIでのリアルタイム進捗表示を可能にする。
- 参照用語のバッチ検索進捗も通知する。

### 10. ライブラリの選定
- DBアクセス (PM側): `github.com/mattn/go-sqlite3` または標準 `database/sql`
- 依存性注入: `github.com/google/wire`
- 並行処理: Go標準 `sync`, `context`
- ステミング: `github.com/kljensen/snowball`（Snowball English Stemmer）

## 関連ドキュメント
- [クラス図](./context_engine_class_diagram.md) ✅
- [シーケンス図](./context_engine_sequence_diagram.md) ✅
- [テスト設計](./context_engine_test_spec.md) ✅
- [要件定義書](../requirements.md)
- [Term Translator Slice 仕様書](../TermTranslatorSlice/spec.md)
- [PersonaGen Slice 仕様書](../PersonaGenSlice/spec.md)
- [Summary Generator Slice 仕様書](../SummaryGeneratorSlice/spec.md)
- [LLMクライアントインターフェース](../LLMClient/llm_client_interface.md)
- [Config Store 仕様書](../ConfigStoreSlice/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
