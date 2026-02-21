# 本文翻訳 (Pass 2 Translator Slice) 仕様書

## 概要
Context Engine Sliceが構築した `TranslationRequest` のリストを受け取り、LLMによる本文翻訳を実行し、翻訳結果を出力する機能である。
本機能は2-Pass Systemにおける **Pass 2: 本文翻訳 (Main Translation)** の中核を担い、Pass 1で構築された用語辞書・ペルソナ・要約等のコンテキスト情報を活用して高品質な翻訳を生成する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「プロンプト構築」「HTMLタグ前処理/後処理」「LLM呼び出し制御」「リトライ制御」「翻訳結果の逐次保存」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `translator.py` が翻訳実行エンジンとして、プロンプト構築・LLM呼び出し・リトライ・HTMLタグ処理・バッチ並列実行・逐次保存を担っている。
- Go v2ではこれらの責務を本 Pass 2 Translator Slice として凝集させ、Context Engine Sliceが構築した `TranslationRequest` を入力として受け取る明確な境界を設ける。
- 要件定義書 §3.3「翻訳エンジン」に基づき、レコード種別に応じたプロンプト生成、タグ保護、リトライ/リカバリ、逐次保存を実現する。

## スコープ
### 本Sliceが担う責務
1. **プロンプト構築**: `TranslationRequest` のレコードタイプとコンテキスト情報に基づき、レコード種別ごとに最適なシステムプロンプトとユーザープロンプトを動的に生成する。
2. **HTMLタグ前処理/後処理**: ゲーム内特有のタグ（`<font>`, `<alias>` 等）を翻訳前に抽象化プレースホルダーに置換し、翻訳後に復元する。
3. **LLM翻訳の実行**: LLMクライアントインターフェースを通じて翻訳を実行する。
4. **リトライ制御**: AIの応答エラーやフォーマット不正時に、指数バックオフによる自動リトライを行う。
5. **強制翻訳の適用**: `ForcedTranslation` が設定されたリクエストはLLM呼び出しをスキップし、辞書の訳をそのまま採用する。
6. **書籍の長文分割翻訳（Chunking）**: 書籍テキストをチャンク分割し、各チャンクを個別に翻訳して結合する。
7. **翻訳結果の逐次保存**: 翻訳完了ごとに結果を保存し、不慮の停止時のデータ損失を最小限に抑える。
8. **差分更新（Resume）**: 既に翻訳済みの出力ファイルが存在する場合、成功済みレコードをスキップし、未翻訳・失敗レコードのみを処理する。
9. **バッチ並列翻訳**: Goroutineによる並列翻訳で処理を高速化する。

### 本Sliceの責務外
- 翻訳リクエストの構築（Context Engine Sliceの責務）
- 用語翻訳（Term Translator Sliceの責務）
- NPCペルソナ生成（Persona Generator Sliceの責務）
- 会話・クエスト要約の生成（Summary Generator Sliceの責務）
- 最終出力ファイルの生成（Export Sliceの責務）

## 要件

### 1. 翻訳結果のデータ構造

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
}
```

**ステータス定義**:
| Status | 説明 |
| :--- | :--- |
| `success` | LLM翻訳が成功した |
| `failed` | リトライ上限に達しても翻訳が失敗した |
| `skipped` | 翻訳対象外（既に日本語、空テキスト等）としてスキップされた |
| `cached` | 差分更新により既存の翻訳結果を再利用した |

### 2. プロンプト構築

`TranslationRequest` のレコードタイプに基づき、レコード種別ごとに最適なシステムプロンプトとユーザープロンプトを動的に生成する。

#### 2.1 プロンプトテンプレート構造

システムプロンプトは以下の共通構造を持つ:
```
{base}           -- 翻訳者としての基本指示
{specific}       -- レコード種別固有の指示
{prohibition}    -- 禁止事項（原文の改変禁止、タグ保護等）
{format}         -- 出力フォーマット指示
```

**プロンプトテンプレートのカスタマイズ**:
- レコード種別ごとのシステムプロンプトテンプレートは Config Store に永続化され、UI上で閲覧・編集可能とする（要件定義書 §3.4）。
- 本Sliceは翻訳実行時に Config Store からテンプレートを取得し、動的にプロンプトを構築する。
- テンプレートが未設定の場合はデフォルトテンプレートにフォールバックする。

#### 2.2 レコード種別ごとのプロンプト構築

**会話系（INFO NAM1, INFO RNAM, DIAL FULL）**:
- システムプロンプトに以下のコンテキストを含める:
  - 話者情報（名前・性別・種族・ボイスタイプ・口調指示）
  - ペルソナテキスト（存在する場合）
  - 直前のセリフ（Previous Line）
  - トピック名
  - 会話要約（Dialogue Summary）
  - Mod説明文
  - 参照用語リスト
- プレイヤー発言（`INFO RNAM`, `DIAL FULL`）の場合、話者を "Player" とし、プレイヤー口調指示を適用する。

**クエスト系（QUST FULL, QUST CNAM, QUST NNAM）**:
- システムプロンプトに以下のコンテキストを含める:
  - クエスト名
  - クエスト要約（Quest Summary）
  - ステージ/目標インデックス
  - Mod説明文
  - 参照用語リスト

**アイテム系（{Type} DESC）**:
- システムプロンプトに以下のコンテキストを含める:
  - アイテム種別ヒント（TypeHint）
  - Mod説明文
  - 参照用語リスト

**書籍系（BOOK DESC）**:
- アイテム系と同様のコンテキストに加え、書籍翻訳固有の指示（文体維持、段落構造保持等）を含める。

**汎用（上記以外）**:
- 基本的なシステムプロンプトとMod説明文・参照用語リストのみを含める。

**`PromptBuilder` インターフェース**:
```go
// PromptBuilder はTranslationRequestからシステムプロンプトとユーザープロンプトを構築する。
type PromptBuilder interface {
    Build(request TranslationRequest) (systemPrompt string, userPrompt string, err error)
}
```

### 3. HTMLタグ前処理/後処理

ゲーム内特有のHTMLタグを翻訳前に抽象化し、翻訳後に復元する。

**前処理（Preprocess）**:
1. 翻訳対象テキスト内のHTMLタグ（`<font ...>`, `</font>`, `<alias ...>`, `<br>` 等）を検出する。
2. 各タグを一意のプレースホルダー（例: `[TAG_1]`, `[TAG_2]`）に置換する。
3. タグとプレースホルダーの対応マップ（`tag_map`）を保持する。
4. テキストの先頭・末尾に位置するタグ（エッジタグ）は、翻訳品質への影響が少ないためスキップ可能とする（Config定義）。

**後処理（Postprocess）**:
1. 翻訳結果テキスト内のプレースホルダーを `tag_map` に基づいて元のHTMLタグに復元する。
2. 復元後、原文に存在しないタグが翻訳結果に含まれていないかバリデーションする。
3. 原文に存在しないタグが検出された場合、`TagHallucinationError` としてリトライ対象とする。

**`TagProcessor` インターフェース**:
```go
// TagProcessor はHTMLタグの前処理/後処理を行う。
type TagProcessor interface {
    Preprocess(text string) (processedText string, tagMap map[string]string)
    Postprocess(text string, tagMap map[string]string) string
    Validate(translatedText string, tagMap map[string]string) error
}
```

### 4. LLM翻訳の実行

**単一リクエストの翻訳フロー**:
1. **バリデーション**: ソーステキストが空または空白のみの場合、`skipped` として返却する。
2. **強制翻訳チェック**: `ForcedTranslation` が非nilの場合、LLM呼び出しをスキップし、辞書の訳を `success` として返却する。
3. **HTMLタグ前処理**: §3 に従いタグを抽象化する。
4. **最大トークン数の計算**: ソーステキストのトークン数に基づき、最大生成トークン数を動的に計算する。
   - 計算式: `max_tokens = max(token_count(source_text) * 2.5, 100)`
5. **プロンプト構築**: §2 に従いシステムプロンプトとユーザープロンプトを生成する。
6. **LLM API呼び出し**: LLMクライアントインターフェースを通じて翻訳を実行する。
7. **レスポンスパース**: LLMの応答から翻訳テキストを抽出する。
8. **HTMLタグ後処理**: §3 に従いタグを復元し、バリデーションする。
9. **翻訳結果の構築**: `TranslationResult` を生成して返却する。

**`Translator` インターフェース**:
```go
// Translator は単一のTranslationRequestを翻訳する。
type Translator interface {
    Translate(ctx context.Context, request TranslationRequest) (TranslationResult, error)
}
```

### 5. リトライ制御

LLM呼び出しの失敗時に、指数バックオフによる自動リトライを行う。

**リトライ設定（Config定義）**:
| パラメータ | デフォルト値 | 説明 |
| :--- | :--- | :--- |
| `max_retries` | 5 | 最大リトライ回数 |
| `base_delay_seconds` | 1.0 | 基本待機時間（秒） |
| `max_delay_seconds` | 10.0 | 最大待機時間（秒） |
| `exponential_base` | 2.0 | 指数バックオフの底 |

**リトライ対象エラー**:
- LLM APIの一時的なエラー（タイムアウト、レート制限、サーバーエラー）
- レスポンスパース失敗（翻訳テキストの抽出不可）
- タグハルシネーションエラー（`TagHallucinationError`）

**待機時間の計算**:
```
delay = min(base_delay_seconds * exponential_base ^ attempt, max_delay_seconds)
```

**リトライ不可エラー**:
- コンテキストのキャンセル（`context.Canceled`）
- 致命的なAPI認証エラー

### 6. 書籍の長文分割翻訳（Chunking）

書籍テキスト（`BOOK DESC`）が長文の場合、チャンク分割して個別に翻訳し、結果を結合する。

**分割ルール**:
1. テキストをHTMLタグ構造を維持しつつ、段落（`<p>`, `<br>`, 改行）単位で分割する。
2. 各チャンクのトークン数が閾値（Config定義、デフォルト: 1500トークン）以下になるよう調整する。
3. 分割不可能な長大な段落は、文単位でさらに分割する。

**翻訳フロー**:
1. 書籍テキストをチャンク分割する。
2. 各チャンクを個別に翻訳する（§4 の単一リクエスト翻訳フローに従う）。
3. 全チャンクの翻訳結果を元の順序で結合する。
4. いずれかのチャンクが翻訳失敗した場合、書籍全体を `failed` とする。

**`BookChunker` インターフェース**:
```go
// BookChunker は書籍テキストをチャンク分割する。
type BookChunker interface {
    Chunk(text string, maxTokensPerChunk int) []string
}
```

### 7. バッチ並列翻訳

複数の `TranslationRequest` を Goroutine で並列翻訳する。

**並列制御**:
- 並列度（ワーカー数）はConfig定義する（デフォルト: 4）。
- `semaphore` パターンまたは `errgroup` を使用して並列度を制御する。
- 各ワーカーは `context.Context` を受け取り、キャンセルシグナルに応答する。

**実行フロー**:
1. `TranslationRequest` のリストを受け取る。
2. 差分更新チェック（§8）により、翻訳済みリクエストを `cached` としてスキップする。
3. 未翻訳リクエストを並列ワーカーに分配する。
4. 各ワーカーが §4 の翻訳フローを実行する。
5. 翻訳完了ごとに結果を逐次保存する（§9）。
6. 全リクエストの処理完了後、`TranslationResult` のリストを返却する。

**`BatchTranslator` インターフェース**:
```go
// BatchTranslator は複数のTranslationRequestをバッチ翻訳する。
type BatchTranslator interface {
    TranslateBatch(
        ctx context.Context,
        requests []TranslationRequest,
        config BatchConfig,
    ) ([]TranslationResult, error)
}
```

**`BatchConfig` 構造体**:
```go
// BatchConfig はバッチ翻訳の実行設定を保持する。
type BatchConfig struct {
    MaxWorkers      int     // 並列ワーカー数
    TimeoutSeconds  float64 // 単一リクエストのタイムアウト（秒）
    MaxTokens       int     // 最大トークン数（0の場合は動的計算）
    OutputBaseDir   string  // 出力ベースディレクトリ
    PluginName      string  // プラグイン名（逐次保存用）
}
```

### 8. 差分更新（Resume）

既に翻訳済みの出力ファイルが存在する場合、成功済みレコードをスキップする。

**Resumeフロー**:
1. 出力ベースディレクトリとプラグイン名から、既存の翻訳結果ファイルを検索する。
2. 既存ファイルから翻訳済みレコードのキー（FormID + RecordType + SourcePlugin）を抽出する。
3. 新規リクエストのキーと照合し、一致するリクエストは `cached` ステータスで既存の翻訳結果を返却する。
4. 一致しないリクエストのみをLLM翻訳の対象とする。

**キー生成ルール**:
```
request_key = "{source_plugin}|{id}|{record_type}" + (index != nil ? "|{index}" : "")
```

### 9. 翻訳結果の逐次保存

翻訳完了ごとに結果をファイルに保存し、不慮の停止時のデータ損失を最小限に抑える。

**保存ルール**:
1. 翻訳結果はソースプラグイン単位でJSONファイルに保存する。
2. 保存先: `{output_base_dir}/{plugin_name}/` 配下。
3. 翻訳が1件完了するごとに、該当プラグインの出力ファイルを更新する。
4. ファイル書き込みは排他制御（Mutex）で保護する。
5. 書き込み失敗時はエラーログを出力するが、翻訳処理自体は継続する。

**出力フォーマット**:
```json
[
  {
    "form_id": "0x001234|Skyrim.esm",
    "editor_id": "DialogTopic01",
    "type": "INFO",
    "original": "I took an arrow in the knee.",
    "string": "昔はお前のような冒険者だったのだが、膝に矢を受けてしまってな。"
  }
]
```

**`ResultWriter` インターフェース**:
```go
// ResultWriter は翻訳結果を逐次保存する。
type ResultWriter interface {
    Write(result TranslationResult) error
    Flush() error
}
```

### 10. 翻訳検証（Translation Verification）

翻訳結果の品質を検証するオプション機能。

**検証項目**:
1. **タグ整合性**: 原文のHTMLタグが翻訳結果に正しく復元されているか。
2. **空翻訳チェック**: 翻訳結果が空または空白のみでないか。
3. **原文コピーチェック**: 翻訳結果が原文と完全一致していないか（翻訳されていない可能性）。

**`TranslationVerifier` インターフェース**:
```go
// TranslationVerifier は翻訳結果の品質を検証する。
type TranslationVerifier interface {
    Verify(sourceText string, translatedText string, tagMap map[string]string) error
}
```

### 11. 進捗通知
- 翻訳の進捗（完了数/総数、成功数/失敗数/スキップ数）をコールバックまたはチャネル経由でProcess Managerに通知し、UIでのリアルタイム進捗表示を可能にする。
- プラグイン単位の進捗も個別に通知する。

### 12. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- 依存性注入: `github.com/google/wire`
- 並行処理: Go標準 `sync`, `context`, `golang.org/x/sync/errgroup`
- トークンカウント: `github.com/pkoukk/tiktoken-go`（tiktoken Go実装）
- JSON処理: Go標準 `encoding/json`

## 関連ドキュメント
- [クラス図](./pass2_translator_class_diagram.md) ✅
- [シーケンス図](./pass2_translator_sequence_diagram.md) ✅
- [テスト設計](./pass2_translator_test_spec.md) ✅
- [要件定義書](../requirements.md)
- [Context Engine Slice 仕様書](../ContextEngineSlice/spec.md)
- [Term Translator Slice 仕様書](../TermTranslatorSlice/spec.md)
- [PersonaGen Slice 仕様書](../PersonaGenSlice/spec.md)
- [Summary Generator Slice 仕様書](../SummaryGeneratorSlice/spec.md)
- [LLMクライアントインターフェース](../LLMClient/llm_client_interface.md)
- [Config Store 仕様書](../ConfigStoreSlice/spec.md)
