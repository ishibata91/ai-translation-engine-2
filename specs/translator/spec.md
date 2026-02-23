# 本文翻訳 (Pass 2 Translator Slice) 仕様書

## 概要
Lore Sliceが構築した `TranslationRequest` のリストを受け取り、LLMによる本文翻訳を実行し、翻訳結果を出力する機能である。
本機能は2-Pass Systemにおける **Pass 2: 本文翻訳 (Main Translation)** の中核を担い、Pass 1で構築された用語辞書・ペルソナ・要約等のコンテキスト情報を活用して高品質な翻訳を生成する。

当機能は、バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト（ジョブ）生成」と「結果保存」の2段階に分離する **2フェーズモデル（提案/保存モデル）** を採用する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「プロンプト構築」「HTMLタグ前処理/後処理」「リトライ制御」「翻訳結果の逐次保存」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `translator.py` が翻訳実行エンジンとして、プロンプト構築・LLM呼び出し・リトライ・HTMLタグ処理・バッチ並列実行・逐次保存を担っている。
- Go v2ではこれらの責務を本 Pass 2 Translator Slice として凝集させ、Lore Sliceが構築した `TranslationRequest` を入力として受け取る明確な境界を設ける。
- **2フェーズモデルへの移行**: LLMへの通信制御をスライスから分離し、Job Queue / Pipeline へ委譲することで、ネットワーク分断やバッチAPI待機に対する堅牢性を確保する。
- 要件定義書 §3.3「翻訳エンジン」に基づき、レコード種別に応じたプロンプト生成、タグ保護、リトライ/リカバリ、逐次保存を実現する。

## スコープ
### 本Sliceが担う責務
1. **翻訳ジョブの生成 (Phase 1: Propose)**: `TranslationRequest` 群を受け取り、差分更新（Resume）チェック、タグ保護（プレースホルダー化）、プロンプト構築を行い、LLMリクエスト群（ジョブ）を構築して返す。
2. **翻訳結果の保存 (Phase 2: Save)**: LLMからのレスポンス群を受け取り、タグ復元、パース、バリデーション（タグハルシネーションチェック）を行い、**ソースプラグイン単位のJSONファイル**に保存する。
3. **強制翻訳の適用**: `ForcedTranslation` が設定されたリクエストは Phase 1 で即時結果として分離する。
4. **書籍の長文分割翻訳（Chunking）**: 書籍テキストをチャンク分割し、個別のリクエストとしてジョブに含める。

### 本Sliceの責務外
- LLMへの実際のHTTP通信制御（Job Queue / LLM Client の責務）
- 翻訳リクエストの構築（Lore Sliceの責務）
- 用語翻訳（Terminology Sliceの責務）
- NPCペルソナ生成（Personaerator Sliceの責務）
- 会話・クエスト要約の生成（Summary Sliceの責務）

## 要件

### 1. 2フェーズモデル（提案/保存モデル） (Propose/Save Model)
**Reason**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成」と「結果保存」の2段階に分離し、通信制御をインフラ層（JobQueue/Pipeline）へ委譲する。

#### Scenario: 翻訳ジョブの提案 (Phase 1: Propose)
- **WHEN** プロセスマネージャーから `[]TranslationRequest` を受け取った
- **THEN** 既存ファイルから成功済みレコードを検索（Resume）し、未処理分を特定する
- **AND** 未処理分に対し、タグ保護とプロンプト構築を行い、`[]llm.Request` を返す
- **AND** 強制翻訳可能なレコードや `cached` 分は、即時結果として分離して返す
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: 翻訳結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm.Response` が渡された
- **THEN** 各レスポンスから翻訳文を抽出し、タグ復元とバリデーションを行う
- **AND** パースに成功した結果をソースプラグイン単位のJSONファイルに逐次保存する
- **AND** タグ破損等の異常が検出された結果についてはエラーとして記録し、リトライが必要な場合はオーケストレーターに通知する
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: 第2パス翻訳向けデータの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTOを定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

### 3. 翻訳結果のデータ構造

**`TranslationResult` 構造体**:
```go
// TranslationResult は単一レコードの翻訳結果を表す。
type TranslationResult struct {
    ID             string  // FormID
    RecordType     string  // レコードタイプ（例: "INFO NAM1", "QUST CNAM"）
    SourceText     string  // 原文（英語）
    TranslatedText *string // 翻訳結果（日本語）。失敗時はnil。
    Index          *int    // Stage/Objective Index（該当なしの場合はnil）
    Status         string  // 処理状態（"success", "failed", "skipped", "cached"）
    ErrorMessage   *string // エラー時のメッセージ
    SourcePlugin   string  // ソースプラグイン名
    SourceFile     string  // リクエスト発生元のファイル名
    EditorID       *string // Editor ID
    ParentID       *string // 親のFormID
    ParentEditorID *string // 親のEditorID
}
```

### 4. プロンプト構築
レコードタイプに基づき、レコード種別ごとに最適なシステムプロンプトとユーザープロンプトを動的に生成する。
- テンプレートは Config に永続化され、UI上で編集可能。
- 話者属性、会話要約、クエスト要約、参照用語リスト等のコンテキストをプレースホルダーに埋め込む。

### 5. HTMLタグ前処理/後処理
ゲーム内特有のHTMLタグ（`<font>`, `<alias>` 等）を翻訳前に抽象化プレースホルダー（`[TAG_1]` 等）に置換し、翻訳後に復元する。
- **タグハルシネーションチェック**: 復元後、原文に存在しないタグが捏造されていないか、または必要なタグが消失していないかバリデーションする。

### 6. 書籍の長文分割翻訳（Chunking）
書籍テキスト（`BOOK DESC`）をHTML構造を維持しつつ、一定文字数でチャンク分割する。各チャンクを個別のLLMリクエストとしてジョブに含め、保存時に再結合する。

### 7. 差分更新（Resume）
出力ディレクトリ内の既存JSONファイルを読み込み、成功済みレコードをスキップする。

### 8. 翻訳結果の逐次保存
翻訳が完了（またはレスポンスをバッチパース）するごとに、ソースプラグイン単位のJSONファイルに結果を書き込む。

### 9. レコードシグネチャの完全保持 (Preservation of Full Signature)
**Reason**: XML出力の `<REC>` タグ生成において正確なシグネチャ情報が必要となるため。
**Requirement**: 翻訳結果の保存時、レコードの `type` (RecordType) を丸めず（例: `INFO` への短縮などを行わず） `INFO NAM1` のようにフル形式で保持しなければならない。

### 9. ライブラリの選定
- LLMクライアント: `infrastructure/llm` インターフェース（プロジェクト共通）
- 依存性注入: `github.com/google/wire`
- JSON処理: Go標準 `encoding/json`

## 関連ドキュメント
- [クラス図](main_translator_class_diagram.md) ✅
- [シーケンス図](main_translator_sequence_diagram.md) ✅
- [テスト設計](main_translator_test_spec.md) ✅
- [要件定義書](../requirements.md)
- [Lore Slice 仕様書](../lore/spec.md)
- [Terminology Slice 仕様書](../terminology/spec.md)
- [Persona Slice 仕様書](../persona/spec.md)
- [Summary Slice 仕様書](../summary/spec.md)
- [LLMクライアントインターフェース](../llm/llm_interface.md)
- [Config 仕様書](../config/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
