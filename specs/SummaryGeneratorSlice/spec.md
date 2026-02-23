# 要約ジェネレータ (Summary Generator Slice) 仕様書

## 概要
クエストの進行状況および会話の流れを、翻訳用コンテキストとして要約する機能である。
LLMを用いて「クエスト要約（これまでのあらすじ）」と「会話要約（会話の流れ）」を生成し、SQLiteにキャッシュすることで、Pass 2（本文翻訳）時に高品質なコンテキストを提供する。

当機能は、バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト（ジョブ）生成」と「結果保存」の2段階に分離する **2フェーズモデル（提案/保存モデル）** を採用する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「要約テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `slm_summarizer.py` が会話・クエストの要約をLLMで生成し、`summary_cache.py` がサイドカーJSONファイルにキャッシュしている。
- Go v2ではキャッシュ先をSQLiteに移行し、永続化の信頼性と検索性を向上させる。
- **2フェーズモデルへの移行**: LLMへの通信制御をスライスから分離し、Job Queue / Process Manager へ委譲することで、ネットワーク分断やバッチAPI待機に対する堅牢性を確保する。
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
- 会話ツリーの構造解析（Context Engine Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）
- NPCペルソナ生成（Persona Generator Sliceの責務）
- 用語翻訳（Term Translator Sliceの責務）

## 要件

### 1. 2フェーズモデル（提案/保存モデル） (Propose/Save Model)
**Reason**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成」と「結果保存」の2段階に分離し、通信制御をインフラ層（JobQueue/ProcessManager）へ委譲する。

#### Scenario: 要約ジョブの提案 (Phase 1: Propose)
- **WHEN** プロセスマネージャーから `SummaryGeneratorInput` 形式のデータを受け取った
- **THEN** 各項目（会話/クエスト）に対してキャッシュヒット判定を行う
- **AND** キャッシュミスした項目について、LLMプロンプトを構築して `[]llm_client.Request` を返す
- **AND** キャッシュヒットした項目は、即時結果として分離して返す
- **AND** `specs/refactoring_strategy.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: 要約結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから要約テキストを抽出し、SQLiteキャッシュに UPSERT する
- **AND** `specs/refactoring_strategy.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: サマリ生成向けデータの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTOを定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化とサマリ生成
- **WHEN** オーケストレーター層から本スライス専用の入力DTOが提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いてサマリ生成処理を完結できること

### 3. 会話要約の生成

本Sliceは、`DialogueGroupInput` から会話の流れを収集し、LLMで要約する。

**収集ルール**:
1. `DialogueGroupInput` 単位で処理する。
2. 前後の発言テキストを含む会話フローを構築する。

**要約生成**:
1. 1文（500文字以内）の英語要約を生成する。
2. 要約は「誰が誰に何について話しているか」を明確にする。

**システムプロンプト**:
```
You are a Skyrim translation assistant providing effective contextual information.
Summarize the conversation flow in a single English sentence (under 500 characters) that maintains the context.
Clearly identify the subject and concisely describe who is speaking to whom and about what.

Output Example:
- Ulfric is ordering his subordinate to march on Whiterun.
- The innkeeper is telling the player about recent rumors.
```

### 4. クエスト要約の生成

本Sliceは、`QuestInput` から累積的な要約を生成する。

**処理ルール**:
1. `QuestInput.StageTexts` を時系列順に処理する。
2. 各ステージの処理時、それまでの全ステージの記述テキストを入力としてLLMに送信し、「これまでのあらすじ」を生成する。

**要約生成**:
1. 1文（500文字以内）の英語要約を生成する。
2. 要約はクエストの核心的なプロットと翻訳に必要な背景知識に焦点を当てる。

**システムプロンプト**:
```
You are a Skyrim translation assistant providing contextual information.
Summarize the overall quest story based on the provided stage descriptions in one concise English sentence (under 500 characters).
Focus on the core plot and background knowledge necessary for translation.

Example Output:
- Retrieve the Dragonstone and report the secrets of Bleak Falls Barrow to the Jarl.
- Return the Golden Claw to Lucan and help Camille.
```

### 5. キャッシュキー生成とヒット判定

要約の再生成を回避するため、入力テキストに基づくキャッシュキーを生成し、SQLiteに保存する。

**キャッシュキー生成方式**:
1. 要約対象のレコードID（`record_id`）と、要約対象テキストの連結文字列のSHA-256ハッシュを組み合わせる。
2. キャッシュキー形式: `{record_id}|{sha256_hash}`
3. 入力テキストが変更された場合（ハッシュ不一致）、キャッシュは無効とし再生成する。

### 6. 要約の永続化（ソースファイル単位キャッシュ）

生成された要約を**ソースファイル（プラグイン）単位のSQLiteファイル**に保存し、Pass 2で参照可能にする。

**設計方針**:
- 各ソースファイル（例: `Skyrim.esm`）に対して、個別のSQLiteキャッシュファイルを作成する。
- ファイル命名規則: `{source_plugin_name}_summary_cache.db`
- 本Slice内の `SummaryStore` が要約テーブルに対するすべての操作を単独で完結させる。

### 7. 要約DBスキーマ

各ソースファイルのSQLiteファイル内に以下のテーブルを作成する。

#### テーブル: `summaries`
| カラム             | 型                                | 説明                                               |
| :----------------- | :-------------------------------- | :------------------------------------------------- |
| `id`               | INTEGER PRIMARY KEY AUTOINCREMENT | 自動採番ID                                         |
| `record_id`        | TEXT NOT NULL                     | 対象レコードID（DialogueGroup.ID または Quest.ID） |
| `summary_type`     | TEXT NOT NULL                     | 要約種別（`"dialogue"` または `"quest"`）          |
| `cache_key`        | TEXT NOT NULL UNIQUE              | キャッシュキー（`{record_id}\|{sha256_hash}`）     |
| `input_hash`       | TEXT NOT NULL                     | 入力テキストのSHA-256ハッシュ                      |
| `summary_text`     | TEXT NOT NULL                     | 生成された要約テキスト                             |
| `input_line_count` | INTEGER NOT NULL                  | 要約対象の入力行数                                 |
| `created_at`       | DATETIME DEFAULT CURRENT_TIMESTAMP | 作成日時                                           |
| `updated_at`       | DATETIME DEFAULT CURRENT_TIMESTAMP | 更新日時                                           |

### 8. Pass 2での参照

Pass 2（本文翻訳）において、翻訳対象レコードに対応する要約をコンテキストとして参照する。

**参照ルール**:
1. **会話翻訳時**: `record_id` をキーとして、該当ソースファイルの `summaries` テーブルから要約を検索する。
2. **クエスト関連翻訳時**: `record_id` をキーとして要約を検索し、「これまでのあらすじ」として含める。

### 9. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス: `github.com/mattn/go-sqlite3` または `modernc.org/sqlite`
- 依存性注入: `github.com/google/wire`
- ハッシュ: Go標準 `crypto/sha256`

## 関連ドキュメント
- [クラス図](./summary_generator_class_diagram.md)
- [シーケンス図](./summary_generator_sequence_diagram.md)
- [テスト設計](./summary_generator_test_spec.md)
- [要件定義書](../requirements.md)
- [PersonaGen Slice 仕様書](../PersonaGenSlice/spec.md)
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
