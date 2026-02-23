# 要約ジェネレータ (Summary Slice) 仕様書

## 概要
クエストの進行状況および会話の流れを、翻訳用コンテキストとして要約する機能である。
LLMを用いて「クエスト要約（これまでのあらすじ）」と「会話要約（会話の流れ）」を生成し、SQLiteにキャッシュすることで、Pass 2（本文翻訳）時に高品質なコンテキストを提供する。

当機能は、バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト（ジョブ）生成」と「結果保存」の2段階に分離する **2フェーズモデル（提案/保存モデル）** を採用する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「要約テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `slm_summarizer.py` が会話・クエストの要約をLLMで生成し、`summary_cache.py` がサイドカーJSONファイルにキャッシュしている。
- Go v2ではキャッシュ先をSQLiteに移行し、永続化の信頼性と検索性を向上させる。
- **2フェーズモデルへの移行**: LLMへの通信制御をスライスから分離し、Job Queue / Pipeline へ委譲することで、ネットワーク分断やバッチAPI待機に対する堅牢性を確保する。
- 要件定義書 §3.2「クエスト要約」に基づき、クエストステージを時系列順に処理し、過去のステージの翻訳結果を「これまでのあらすじ」として累積的に次ステージのコンテキストに含める。
- 会話ツリーの前後関係を要約することで、翻訳時の文脈理解を向上させる。

## スコープ
### 本Sliceが担う責務
1. **要約ジョブの生成 (Phase 1: Propose)**: `DialogueGroup` や `Quest` のデータを元に、キャッシュヒット判定を行い、未キャッシュの項目についてLLMリクエスト群（ジョブ）を構築して返す。
2. **要約のキャッシュ保存 (Phase 2: Save)**: LLMからのレスポンス群を受け取り、パース・バリデーションを行った上で**ソースファイル単位のSQLiteファイル**に保存する。
3. **要約の取得**: レコードIDに基づき、キャッシュ済みの要約テキストを取得する。
4. **キャッシュヒット判定**: 入力テキストのハッシュに基づき、既存キャッシュの有効性を判定する。

### 本Sliceの責務外
- LLMへの実際のHTTP通信制御（Job Queue / LLM Client の責務）
- 会話ツリーの構造解析（Lore Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）
- NPCペルソナ生成（Personaerator Sliceの責務）
- 用語翻訳（Terminology Sliceの責務）

## 要件

### 1. 2フェーズモデル（提案/保存モデル） (Propose/Save Model)
**Reason**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成」と「結果保存」の2段階に分離し、通信制御をインフラ層（JobQueue/Pipeline）へ委譲する。

#### Scenario: 要約ジョブの提案 (Phase 1: Propose)
- **WHEN** プロセスマネージャーから `SummaryGeneratorInput` 形式のデータを受け取った
- **THEN** 各項目（会話/クエスト）に対してキャッシュヒット判定を行う
- **AND** キャッシュミスした項目について、LLMプロンプトを構築して `[]llm_client.Request` を返す
- **AND** キャッシュヒットした項目は、即時結果として分離して返す
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: 要約結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから要約テキストを抽出し、SQLiteキャッシュに UPSERT する
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: SummaryGeneratorSliceの独立した初期化
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
本スライスは独自の入力DTO（`SummaryGeneratorInput`）を `contract.go` 内に定義し、他スライスのデータ構造に依存してはならない。

#### Scenario: オーケストレーターが入力DTOをSummaryGeneratorSliceに渡す
- **WHEN** オーケストレーターが Parser / ContextEngineSlice の出力から会話・クエストデータを抽出し、`SummaryGeneratorInput` にマッピングして渡す
- **THEN** SummaryGeneratorSlice が外部パッケージの型に依存することなく初期化され、要約生成の準備が整う

### 3. 会話要約の生成

本Sliceは、`DialogueGroupInput` から会話の流れを収集し、LLMで要約する。

**収集ルール**:
1. `DialogueGroupInput` 単位で処理する。
2. 前後の発言テキストを含む会話フローを構築する。

**要約生成**:
1. 1文（500文字以内）の英語要約を生成する。
2. 要約は「誰が誰に何について話しているか」を明確にする。

**システムプロンプト**:
#### Scenario: 会話フローを収集してLLMで要約を生成する
- **WHEN** `DialogueGroup` 内に1件以上の会話行があり、LLMが正常に応答する
- **THEN** 「誰が誰に何について話しているか」を明確にした500文字以内の英語1文要約が返却される

#### Scenario: 会話行が0件の場合はスキップする
- **WHEN** `DialogueGroup` 内の会話行が0件である
- **THEN** LLMは呼び出されず、空のサマリとして処理がスキップされる

### 4. クエスト要約の生成

本Sliceは、`QuestInput` から累積的な要約を生成する。

**処理ルール**:
1. `QuestInput.StageTexts` を時系列順に処理する。
2. 各ステージの処理時、それまでの全ステージの記述テキストを入力としてLLMに送信し、「これまでのあらすじ」を生成する。

**要約生成**:
1. 1文（500文字以内）の英語要約を生成する。
2. 要約はクエストの核心的なプロットと翻訳に必要な背景知識に焦点を当てる。

**システムプロンプト**:
#### Scenario: 複数ステージのクエスト要約を累積生成する
- **WHEN** `Quest` が2件以上のステージを持ち、LLMが正常に応答する
- **THEN** 各ステージの処理で過去すべてのステージ記述が入力として使われ、累積的なあらすじが生成される

#### Scenario: ステージ記述が0件の場合はスキップする
- **WHEN** `Quest.Stages` が空である
- **THEN** LLMは呼び出されず、処理がスキップされる

### 5. キャッシュキー生成とヒット判定

要約の再生成を回避するため、入力テキストに基づくキャッシュキーを生成し、SQLiteに保存する。

#### Scenario: キャッシュヒット時にLLM呼び出しをスキップする
- **WHEN** 同一の `record_id` と入力テキストに対して要約生成が再度リクエストされる
- **THEN** SQLiteのキャッシュから要約テキストが返却され、LLMへのリクエストは生成されない

#### Scenario: キャッシュミス時に要約を生成してキャッシュに保存する
- **WHEN** 該当するキャッシュキーがSQLiteに存在しない
- **THEN** LLMによって要約が生成され、`summaries` テーブルにUPSERTされる

### 6. 要約の永続化（ソースファイル単位キャッシュ）

生成された要約を**ソースファイル（プラグイン）単位のSQLiteファイル**に保存し、Pass 2で参照可能にする。

**設計方針**:
- 各ソースファイル（例: `Skyrim.esm`）に対して、個別のSQLiteキャッシュファイルを作成する。
- ファイル命名規則: `{source_plugin_name}_summary_cache.db`
- 本Slice内の `SummaryStore` が要約テーブルに対するすべての操作を単独で完結させる。

#### Scenario: 初回起動時にsummariesテーブルを自動作成する
- **WHEN** `SummaryStore` が初期化される
- **THEN** 接続先SQLiteに `summaries` テーブルが存在しない場合、`CREATE TABLE IF NOT EXISTS` により自動作成される

#### Scenario: 要約をUPSERTで保存する
- **WHEN** 要約生成が成功した後にストアへの保存が呼ばれる
- **THEN** `summaries` テーブルに対して UPSERT（同一 `cache_key` の場合は上書き）が実行される

### 7. 要約DBスキーマ

各ソースファイルのSQLiteファイル内に以下のテーブルを作成する。

#### テーブル: `summaries`
| カラム             | 型                                 | 説明                                               |
| :----------------- | :--------------------------------- | :------------------------------------------------- |
| `id`               | INTEGER PRIMARY KEY AUTOINCREMENT  | 自動採番ID                                         |
| `record_id`        | TEXT NOT NULL                      | 対象レコードID（DialogueGroup.ID または Quest.ID） |
| `summary_type`     | TEXT NOT NULL                      | 要約種別（`"dialogue"` または `"quest"`）          |
| `cache_key`        | TEXT NOT NULL UNIQUE               | キャッシュキー（`{record_id}\|{sha256_hash}`）     |
| `input_hash`       | TEXT NOT NULL                      | 入力テキストのSHA-256ハッシュ                      |
| `summary_text`     | TEXT NOT NULL                      | 生成された要約テキスト                             |
| `input_line_count` | INTEGER NOT NULL                   | 要約対象の入力行数                                 |
| `created_at`       | DATETIME DEFAULT CURRENT_TIMESTAMP | 作成日時                                           |
| `updated_at`       | DATETIME DEFAULT CURRENT_TIMESTAMP | 更新日時                                           |

### 8. Pass 2での参照

Pass 2（本文翻訳）において、翻訳対象レコードに対応する要約をコンテキストとして参照する。

**参照ルール**:
1. **会話翻訳時**: `record_id` をキーとして、該当ソースファイルの `summaries` テーブルから要約を検索する。
2. **クエスト関連翻訳時**: `record_id` をキーとして要約を検索し、「これまでのあらすじ」として含める。

### 9. 並列要約生成
**Reason**: 多数の要約対象を効率的に処理し、実行時間を短縮するため。
本スライスは複数の `DialogueGroup` および `Quest` の要約を Goroutine で並列処理しなければならない。並列度は Config で設定可能（デフォルト: 10）とする。ただし、同一クエスト内のステージは逐次処理（Index昇順）を維持する。

#### Scenario: 複数DialogueGroupを並列処理する
- **WHEN** 複数の `DialogueGroup` が入力として渡される
- **THEN** 設定された並列度でGoroutineが起動し、各グループの要約が並列に生成される

#### Scenario: クエストのステージはIndex昇順に逐次処理する
- **WHEN** 同一 `Quest` の複数ステージを処理する
- **THEN** ステージは `Index` 昇順に逐次処理され、要約が累積的にビルドされる

### 10. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス: `github.com/mattn/go-sqlite3` または `modernc.org/sqlite`
- 依存性注入: `github.com/google/wire`
- ハッシュ: Go標準 `crypto/sha256`
- 並行処理制御: `golang.org/x/sync/errgroup`

## 関連ドキュメント
- [クラス図](./summary_class_diagram.md)
- [シーケンス図](./summary_sequence_diagram.md)
- [テスト設計](./summary_test_spec.md)
- [要件定義書](../requirements.md)
- [Persona Slice 仕様書](../PersonaGenSlice/spec.md)
- [LLMクライアントインターフェース](../LLMClient/llm_client_interface.md)
- [Config 仕様書](../ConfigStoreSlice/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。

#### Scenario: 要約生成開始時のEntryログ出力
- **WHEN** 要約生成処理が開始される
- **THEN** `trace_id`・`span_id`・入力件数を含む DEBUG レベルの Entry ログが JSON 形式で出力される

#### Scenario: 要約生成完了時のExitログ出力
- **WHEN** 要約生成処理が完了する
- **THEN** 生成件数・スキップ件数・経過時間を含む DEBUG レベルの Exit ログが JSON 形式で出力される

3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
