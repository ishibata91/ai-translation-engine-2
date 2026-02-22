# 要約ジェネレータ (Summary Generator Slice) 仕様書

## 概要
クエストの進行状況および会話の流れを、翻訳用コンテキストとして要約する機能である。
LLMを用いて「クエスト要約（これまでのあらすじ）」と「会話要約（会話の流れ）」を生成し、SQLiteにキャッシュすることで、Pass 2（本文翻訳）時に高品質なコンテキストを提供する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「要約テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `slm_summarizer.py` が会話・クエストの要約をLLMで生成し、`summary_cache.py` がサイドカーJSONファイルにキャッシュしている。
- Go v2ではキャッシュ先をSQLiteに移行し、永続化の信頼性と検索性を向上させる。
- 要件定義書 §3.2「クエスト要約」に基づき、クエストステージを時系列順に処理し、過去のステージの翻訳結果を「これまでのあらすじ」として累積的に次ステージのコンテキストに含める。
- 会話ツリーの前後関係を要約することで、翻訳時の文脈理解を向上させる。

## スコープ
### 本Sliceが担う責務
1. **会話要約の生成**: `DialogueGroup` 内の会話の流れ（`PreviousID` チェーンを含む）をLLMで要約する。
2. **クエスト要約の生成**: `Quest` のステージ記述を時系列順に処理し、累積的な「これまでのあらすじ」をLLMで生成する。
3. **要約のキャッシュ（SQLite永続化）**: 生成された要約を**ソースファイル単位のSQLiteファイル**に保存し、同一入力に対する再生成を回避する。
4. **キャッシュヒット判定**: 入力テキストのハッシュに基づき、既存キャッシュの有効性を判定する。

### 本Sliceの責務外
- 会話ツリーの構造解析（Context Engine Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）
- NPCペルソナ生成（Persona Generator Sliceの責務）
- 用語翻訳（Term Translator Sliceの責務）

## 要件

### 独立性: サマリ生成向けデータの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTOを定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化とサマリ生成
- **WHEN** オーケストレーター層から本スライス専用の入力DTOが提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いてサマリ生成処理を完結できること

### 1. 会話要約の生成

本Sliceは、`ExtractedData` に含まれる `DialogueGroup` / `DialogueResponse` から会話の流れを収集し、LLMで要約する。

**収集ルール**:
1. `DialogueGroup` 単位で処理する。各 `DialogueGroup` 内の `Responses` を `Order` 昇順でソートし、会話の流れを構築する。
2. `PreviousID` が設定されている `DialogueResponse` については、参照先の発言テキストを先行コンテキストとして含める。
3. `DialogueGroup.PlayerText` が存在する場合、プレイヤーの発言として会話フローに含める。

**要約生成**:
1. 収集した会話テキストをLLMに送信し、1文（500文字以内）の英語要約を生成する。
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

**ユーザープロンプト構築**:
```
Summarize the following conversation:
- <line 1>
- <line 2>
- ...
```

### 2. クエスト要約の生成

本Sliceは、`ExtractedData` に含まれる `Quest` のステージ記述を時系列順に処理し、累積的な要約を生成する。

**処理ルール**:
1. `Quest.Stages` を `Index` 昇順でソートし、時系列順に処理する。
2. 各ステージの処理時、それまでの全ステージの記述テキストを入力としてLLMに送信し、「これまでのあらすじ」を生成する。
3. 生成された要約は次ステージの翻訳コンテキストとして累積的に使用される。

**要約生成**:
1. ステージ記述テキスト一覧をLLMに送信し、1文（500文字以内）の英語要約を生成する。
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

**ユーザープロンプト構築**:
```
Summarize the overall quest based on the descriptions of these quest stages:
- <stage 1 text>
- <stage 2 text>
- ...
```

### 3. キャッシュキー生成とヒット判定

要約の再生成を回避するため、入力テキストに基づくキャッシュキーを生成し、SQLiteに保存する。

**キャッシュキー生成方式**:
1. 要約対象のレコードID（`DialogueGroup.ID` または `Quest.ID`）と、要約対象テキストの連結文字列のSHA-256ハッシュを組み合わせる。
2. キャッシュキー形式: `{record_id}|{sha256_hash}`
3. 入力テキストが変更された場合（ハッシュ不一致）、キャッシュは無効とし再生成する。

**キャッシュヒット判定フロー**:
1. レコードIDとハッシュでSQLiteを検索する。
2. 一致するレコードが存在すれば、保存済みの要約テキストを返却する（LLM呼び出しをスキップ）。
3. 一致しなければ、LLMで要約を生成し、結果をSQLiteに保存する。

### 4. 要約の永続化（ソースファイル単位キャッシュ）

生成された要約を**ソースファイル（プラグイン）単位のSQLiteファイル**に保存し、Pass 2で参照可能にする。

**ソースファイル単位キャッシュの設計方針**:
- 各ソースファイル（例: `Skyrim.esm`, `Dawnguard.esm`, `MyMod.esp`）に対して、個別のSQLiteキャッシュファイルを作成する。
- ファイル命名規則: `{source_plugin_name}_summary_cache.db`（例: `Skyrim.esm_summary_cache.db`）
- キャッシュファイルはConfig定義のキャッシュディレクトリ配下に格納する。
- これにより、Mod単位でのキャッシュ管理（削除・再生成・配布）が容易になる。

**永続化ルール**:
1. プロセスマネージャーがソースファイル名に基づいて `*sql.DB` を生成し、DIで渡す。
2. 本Slice内の `SummaryStore` が要約テーブルに対するすべての操作（テーブル生成・INSERT/UPSERT・SELECT）を単独で完結させる。
3. 同一キャッシュキーに対する再生成時は UPSERT（既存レコードの上書き）とする。
4. 各ソースファイルのSQLiteには `summaries` テーブルのみが含まれる（他Sliceのテーブルとは同居しない）。

**DB接続管理**:
- `SummaryStore` はソースファイル名を受け取り、対応するSQLiteファイルへの接続を管理する。
- 複数ソースファイルの処理時は、ソースファイルごとに `SummaryStore` インスタンスを生成するか、内部で接続を切り替える。

### 5. 要約DBスキーマ

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
| `created_at`       | DATETIME                          | 作成日時                                           |
| `updated_at`       | DATETIME                          | 更新日時                                           |

> **注**: `source_plugin` カラムは不要。DBファイル自体がソースファイルに対応するため、テーブル内に冗長な情報を持たない。

**インデックス**:
- `idx_summaries_cache_key` ON `summaries(cache_key)` — キャッシュヒット検索用
- `idx_summaries_record_type` ON `summaries(record_id, summary_type)` — レコード別検索用

**DDL**:
```sql
CREATE TABLE IF NOT EXISTS summaries (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id        TEXT NOT NULL,
    summary_type     TEXT NOT NULL CHECK(summary_type IN ('dialogue', 'quest')),
    cache_key        TEXT NOT NULL UNIQUE,
    input_hash       TEXT NOT NULL,
    summary_text     TEXT NOT NULL,
    input_line_count INTEGER NOT NULL,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_summaries_cache_key ON summaries(cache_key);
CREATE INDEX IF NOT EXISTS idx_summaries_record_type ON summaries(record_id, summary_type);
```

### 6. Pass 2での参照

Pass 2（本文翻訳）において、翻訳対象レコードに対応する要約をコンテキストとして参照する。

**参照ルール**:
1. **会話翻訳時**: 翻訳対象の `DialogueGroup.ID` をキーとして、該当ソースファイルの `summaries` テーブルから `summary_type = 'dialogue'` の要約を検索し、LLMプロンプトのコンテキストセクションに挿入する。
2. **クエスト関連翻訳時**: `DialogueGroup.QuestID` または `Quest.ID` をキーとして `summary_type = 'quest'` の要約を検索し、「これまでのあらすじ」としてコンテキストに含める。
3. 要約が存在しない場合、要約なしで翻訳を実行する（フォールバック）。

### 7. 実行制御

**LLM呼び出し**:
- LLMクライアントインターフェース（`infrastructure/llm_client`）を通じて実行する。
- リトライ（指数バックオフ）とタイムアウト制御を備える。
- `max_tokens`: 200（要約は簡潔であるため）
- `temperature`: 0.3（安定した出力のため）

**並列処理**:
- 複数の要約生成を Goroutine で並列実行する。並列度はConfig定義する（デフォルト: 10）。
- 会話要約とクエスト要約は独立して並列処理可能。

**スキップ条件**:
- 入力テキストが空（会話行が0件、ステージ記述が0件）の場合、要約生成をスキップする。
- キャッシュヒットした場合、LLM呼び出しをスキップする。

### 8. 進捗通知
- 要約生成の進捗（完了件数/対象総件数）をコールバックまたはチャネル経由でProcess Managerに通知し、UIでのリアルタイム進捗表示を可能にする。
- 会話要約とクエスト要約の進捗は個別に通知する。

### 9. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス (PM側): `github.com/mattn/go-sqlite3` または標準 `database/sql`
- 依存性注入: `github.com/google/wire`
- 並行処理: Go標準 `sync`, `context`
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
