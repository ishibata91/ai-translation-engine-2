# コンテキストエンジン (Lore Slice) 仕様書

## 概要
翻訳対象レコードに対して、AIが高品質な翻訳を行うために必要な「文脈情報（コンテキスト）」を構築する機能である。
会話ツリーの解析、話者プロファイリング、参照用語の検索、および各レコードタイプに応じた翻訳リクエストの組み立てを担い、Pass 2（本文翻訳）Sliceに対して構造化された `TranslationRequest` のリストを提供する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「翻訳リクエストの構築ロジック」「会話ツリー解析」「話者プロファイリング」「参照用語検索」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `context_builder.py` が2,300行超の巨大モジュールとして、用語翻訳リクエスト生成（Pass 1）と本文翻訳リクエスト生成（Pass 2）の両方を担っている。
- Go v2では責務を明確に分離し、用語翻訳リクエスト生成は Terminology Slice に、本文翻訳リクエスト生成は本 Lore Slice に分割する。
- **JobQueue連携**: 本Sliceは `TranslationRequest` 群を構築するまでを担当し、実際のLLM通信は Pass 2 Translator Slice と JobQueue が担う。
- 要件定義書 §3.2「コンテキスト構築」に基づき、会話ツリー解析・話者プロファイリング・クエスト要約参照・用語検索を統合した構造化コンテキストを構築する。

## スコープ
### 本Sliceが担う責務
1. **会話ツリー解析**: `DialogueGroup` / `DialogueResponse` の `PreviousID` チェーンをトラバースし、直前の発言（Previous Line）を特定する。
2. **話者プロファイリング**: NPCの種族・性別・声タイプからの口調推定、およびペルソナ（Personaerator Slice生成）の参照。
3. **参照用語検索**: 翻訳対象テキストからキーワードを抽出し、辞書DBおよびMod用語DBから関連用語を検索する。
4. **翻訳リクエストの構築**: 上記のコンテキスト情報を統合し、レコードタイプ別の `TranslationRequest` を生成する。
5. **要約の参照**: Summary Sliceが生成した会話要約・クエスト要約をSQLiteから取得し、コンテキストに含める。

### 本Sliceの責務外
- 用語翻訳（Terminology Sliceの責務）
- NPCペルソナ生成（Personaerator Sliceの責務）
- 会話・クエスト要約の生成（Summary Sliceの責務）
- LLMによる本文翻訳の実行（Pass 2 Sliceの責務）
- xTranslator XMLからの辞書DB構築（Dictionary Sliceの責務）

## 要件

### 1. JobQueue連携設計
**Reason**: 本文翻訳（Pass 2）は膨大なリクエストが発生し、バッチAPI等での長時間待機が前提となるため、リクエスト構築（Lore）と翻訳実行（Pass 2 Translator）を明確に分離する。

#### Scenario: 翻訳リクエストの一括構築
- **WHEN** プロセスマネージャーから `ContextEngineInput` 形式のデータを受け取った
- **THEN** 全ての対象レコードをスキャンし、コンテキスト解析と辞書検索を行う
- **AND** 全てのリクエストを `TranslationRequest` 形式の配列として構築して返す
- **AND** 強制翻訳可能なレコードについては `ForcedTranslation` フィールドを埋める
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: コンテキストエンジン向けデータの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライス(LoaderSlice等)のDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（`ContextEngineInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化とコンテキスト構築
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`ContextEngineInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いて文脈解析・構築処理を完結できること

### 3. 翻訳リクエストのデータ構造

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

### 4. 会話ツリー解析
`DialogueGroup` および `DialogueResponse` の情報を基に会話ツリーを解析し、直前の発言を特定する。
1. **DAGベースの探索**: `PreviousID` のチェーンをたどり、各会話ノードのリンクを解析する。
2. **循環参照の防止**: 探索の深さ制限（Max Depth: 5）を設ける。

### 5. 話者プロファイリング
NPCの属性とペルソナ（Personaerator Slice生成）を統合して最適な口調指示文を生成し、`SpeakerProfile` を構築する。
1. **ペルソナによるオーバーライド**: `npc_personas` テーブルの固有テキストを最優先。
2. **属性ベースのマッピング**: 種族・クラス・ボイスタイプに基づく指示文の合成。

### 6. 参照用語検索
翻訳対象テキストからキーワードを抽出し、辞書DBおよびMod用語DBから関連用語を検索する。
- **検索戦略**: 貪欲部分一致（Greedy Partial Match）を採用。
- **強制翻訳（Forced Translation）**: ソーステキスト全文が辞書に完全一致する場合、その訳を `ForcedTranslation` に設定する。

### 7. レコードタイプ別の翻訳リクエスト構築
各ドメインモデル（会話、クエスト、アイテム、書籍、魔法、メッセージ）から、適切なコンテキストを付与した `TranslationRequest` を生成する。
- **書籍の長文分割（Chunking）**: HTMLタグ構造を維持しつつ、一定文字数（1500文字）でチャンク分割し、それぞれリクエスト化する。

### 8. メインインターフェース

**`Lore` インターフェース**:
```go
// Lore は入力データから翻訳リクエストのリストを構築する。
type Lore interface {
    BuildTranslationRequests(
        ctx context.Context,
        input ContextEngineInput,
        config ContextEngineConfig,
    ) ([]TranslationRequest, error)
}
```

### 9. ライブラリの選定
- DBアクセス: `github.com/mattn/go-sqlite3` または `modernc.org/sqlite`
- 依存性注入: `github.com/google/wire`
- ステミング: `github.com/kljensen/snowball`（Snowball English Stemmer）

## 関連ドキュメント
- [クラス図](./lore_class_diagram.md) ✅
- [シーケンス図](./lore_sequence_diagram.md) ✅
- [テスト設計](./lore_test_spec.md) ✅
- [要件定義書](../requirements.md)
- [Terminology Slice 仕様書](../TermTranslatorSlice/spec.md)
- [Persona Slice 仕様書](../PersonaGenSlice/spec.md)
- [Summary Slice 仕様書](../SummaryGeneratorSlice/spec.md)
- [LLMクライアントインターフェース](../LLMClient/llm_client_interface.md)
- [Config 仕様書](../ConfigStoreSlice/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
