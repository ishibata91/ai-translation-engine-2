# NPCペルソナ生成 (NPC Persona Generator Slice) 仕様書

## 概要
NPCごとの会話データ（最大100件）を収集し、LLMにリクエストしてそのNPCの性格・口調・背景を要約した「ペルソナ」を自動生成する機能である。
本機能は2-Pass Systemにおける **Pass 1** の一部として実行され、生成されたペルソナはPass 2（本文翻訳）で話者プロファイルとして参照される。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「ペルソナテーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `context_builder.py` が話者の種族・声タイプから口調を推定しているが、NPCの個別の性格や背景は反映されていない。
- 会話データからLLMでペルソナを生成することで、翻訳時の口調一貫性を大幅に向上させる。
- コンテキスト長の制約があるため、ペルソナ生成前にトークン利用量を事前計算し、超過を防止する必要がある。

## スコープ
### 本Sliceが担う責務
1. **NPC会話データの収集**: ロード済みの `ExtractedData` から、NPCごとに `DIAL/INFO` レコード（`DialogueResponse`）を最大100件まで収集する。
2. **トークン利用量の事前計算**: NPCごとに収集した会話データの想定トークン利用量を計算し、LLMのコンテキスト長制限を超過しないか事前評価する。
3. **LLMによるペルソナ生成**: 会話データとNPC属性（種族・性別・声の種類）をLLMに送信し、ペルソナテキストを生成する。
4. **ペルソナの永続化**: 生成されたペルソナをMod用語DBまたは専用テーブルに保存し、Pass 2で参照可能にする。

### 本Sliceの責務外
- 会話ツリーの構造解析（Context Engine Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）
- 用語翻訳（Term Translator Sliceの責務）

## 要件

### 1. NPC会話データの収集

本Sliceは、`ExtractedData` に含まれる `DialogueGroup` / `DialogueResponse` から、NPCごとの会話データを収集する。

**収集ルール**:
1. `DialogueResponse.SpeakerID` をキーとしてNPCごとにグルーピングする。
2. 各NPCにつき最大 **100件** の `DialogueResponse` を収集する。
3. 100件を超える場合は、**重要度スコアリング**（後述 §1.1）によりスコアの高い上位100件を選別する。
4. `SpeakerID` が nil または空文字の `DialogueResponse` は収集対象外とする。

#### 1.1 重要度スコアリング (Importance Scoring)

NPCの会話データが100件を超える場合、各会話ペア（英語原文・日本語訳）に対して**重要度スコア**を算出し、スコアの高い上位100件（日本語テキスト）をLLMペルソナ生成の入力として選別する。

```
Raw Data (英/日の会話ペア)
  ├── 固有名詞スコアリング → English Noun Hit Count
  ├── 感情スコアリング     → Emotion Indicator Count
  │     ├── Punctuation (!, ?) & ALL CAPS
  │     └── Negative/Strong Words List
  └── Score Aggregation → Top 100 Sentences (Japanese) → LLM Persona Generation
```

**スコアリング要素**:

1. **固有名詞スコア (Proper Noun Score)**:
   - 英語原文に対して、Term Translator Sliceで翻訳済みの固有名詞（NPC名・地名・アイテム名等）を **Word Boundary マッチ (`\b`)** で検索する。
   - ヒットした固有名詞の件数を `proper_noun_hits` とする。
   - 固有名詞を多く含む会話は、NPCの役割・関係性を示す重要な文脈であるため高スコアとする。

2. **感情スコア (Emotion Score)**:
   - 英語原文に対して、以下の感情指標を検出し、合計を `emotion_indicators` とする:
     - **句読点指標**: `!`（感嘆符）および `?`（疑問符）の出現回数。
     - **全大文字指標**: 全大文字の単語（例: `"STOP"`, `"NEVER"`）の出現回数。ALL CAPSはゲーム内での強調・叫びを示す。
     - **ネガティブ/強い単語リスト**: 事前定義された強い感情を示す単語リスト（例: `"die"`, `"kill"`, `"betray"`, `"love"`, `"hate"`, `"fool"`, `"damn"`, `"curse"` 等）とのマッチ回数。単語リストはConfig定義とする。

3. **スコア集計 (Score Aggregation)**:
   - 各会話の重要度スコアを以下の式で算出する:
     ```
     importance_score = proper_noun_hits * W_noun + emotion_indicators * W_emotion + base_priority
     ```
   - `W_noun`, `W_emotion` は重み係数（Config定義、デフォルト: `W_noun=2`, `W_emotion=1`）。
   - `base_priority` はカテゴリ優先度:
     - クエスト関連の会話（`DialogueGroup.QuestID` が非nil）: `+10`
     - サービス分岐の会話（`DialogueGroup.IsServicesBranch == true`）: `+0`
     - その他の会話: `+5`
   - 同一スコアの場合は `DialogueResponse.Order` の昇順で順位を決定する。

4. **選別結果**:
   - スコア降順でソートし、上位100件の**日本語テキスト**をLLMペルソナ生成の入力とする。
   - 日本語テキストが存在しない会話ペアは英語原文をフォールバックとして使用する。

**`ImportanceScorer` インターフェース**:
```go
// ImportanceScorer は会話ペアの重要度スコアを算出する。
type ImportanceScorer interface {
    Score(englishText string, questID *string, isServicesBranch bool) int
}
```

**NPC属性の取得**:
- `SpeakerID` に対応する `NPC` モデルから種族 (`Race`)、性別 (`Sex`)、声の種類 (`Voice`) を取得する。
- NPC属性が取得できない場合でも、会話データのみでペルソナ生成を試みる。

### 2. トークン利用量の事前計算

ペルソナ生成リクエストの送信前に、NPCごとの想定トークン利用量を計算し、LLMのコンテキストウィンドウ超過を防止する。

**計算方式**:
1. **入力トークン数の推定**: 収集した会話テキストの総文字数に基づき、想定トークン数を算出する。
   - 英語テキスト: `文字数 / 4`（概算）
   - システムプロンプト: 固定オーバーヘッドとして加算（Config定義）
2. **出力トークン数の推定**: ペルソナ出力の想定最大トークン数をConfig定義する（デフォルト: 500トークン）。
3. **合計トークン数**: `入力トークン数 + 出力トークン数`

**コンテキスト長評価**:
1. 合計トークン数がLLMのコンテキストウィンドウ上限（Config定義）を超過する場合:
   - 会話データを優先度順に削減し、上限内に収まるよう調整する。
   - 削減後も最低 **10件** の会話データを確保できない場合は、ユーザーへ警告を通知する。
2. 評価結果（NPC名、会話件数、推定トークン数、超過有無）をログ出力する。

**`TokenEstimator` インターフェース**:
```go
// TokenEstimator はテキストの想定トークン数を推定する。
type TokenEstimator interface {
    Estimate(text string) int
}
```

### 3. LLMによるペルソナ生成

収集した会話データとNPC属性をLLMに送信し、ペルソナテキストを生成する。

**リクエスト構築**:
1. システムプロンプトには以下を含める:
   - ペルソナ生成の目的（Skyrim Modの翻訳における口調決定のため）
   - 出力フォーマットの指定（性格特性・話し方の癖・背景設定の3セクション）
2. ユーザープロンプトには以下を含める:
   - NPC属性（種族・性別・声の種類）※取得可能な場合
   - 収集した会話テキスト一覧（最大100件）

**出力フォーマット**:
```
性格特性: <NPCの性格を1-2文で要約>
話し方の癖: <口調・語尾・特徴的な表現を1-2文で要約>
背景設定: <NPCの立場・役割・背景を1-2文で要約>
```

**実行制御**:
- LLMクライアントインターフェース（`infrastructure/llm_client`）を通じて実行する。
- リトライ（指数バックオフ）とタイムアウト制御を備える。
- 並列生成（Goroutine）により処理を高速化する。並列度はConfig定義する。
- 会話データが **0件** のNPCはペルソナ生成をスキップする。

### 4. ペルソナの永続化

生成されたペルソナをSQLiteに保存し、Pass 2で参照可能にする。

**永続化ルール**:
1. プロセスマネージャーから `*sql.DB` をDIで受け取る。
2. 本Slice内の `PersonaStore` がペルソナテーブルに対するすべての操作（テーブル生成・INSERT/UPSERT）を単独で完結させる。
3. 同一NPCに対する再生成時は UPSERT（既存レコードの上書き）とする。
4. ペルソナDBは **Mod用語DBと同一のSQLiteファイル** 内に専用テーブルとして格納する。

### 5. ペルソナDBスキーマ

#### テーブル: `npc_personas`
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | 自動採番ID |
| `speaker_id` | TEXT NOT NULL UNIQUE | NPC識別子（SpeakerID） |
| `editor_id` | TEXT | NPCのEditor ID |
| `npc_name` | TEXT | NPC名（`NPC_:FULL` のテキスト） |
| `race` | TEXT | 種族 |
| `sex` | TEXT | 性別 |
| `voice_type` | TEXT | 声の種類 |
| `persona_text` | TEXT NOT NULL | 生成されたペルソナテキスト |
| `dialogue_count` | INTEGER NOT NULL | ペルソナ生成に使用した会話件数 |
| `estimated_tokens` | INTEGER | 推定トークン利用量 |
| `source_plugin` | TEXT | ソースプラグイン名 |
| `created_at` | DATETIME | 作成日時 |
| `updated_at` | DATETIME | 更新日時 |

### 6. Pass 2での参照

Pass 2（本文翻訳）において、翻訳対象の `DialogueResponse` の `SpeakerID` をキーとして `npc_personas` テーブルからペルソナを検索し、LLMプロンプトのコンテキストに含める。

**参照ルール**:
1. `SpeakerID` で完全一致検索する。
2. ペルソナが存在する場合、システムプロンプトの話者プロファイルセクションに `persona_text` を挿入する。
3. ペルソナが存在しない場合、従来の種族・声タイプベースの口調推定にフォールバックする。

### 7. 進捗通知
- ペルソナ生成の進捗（完了NPC数/対象NPC総数）をコールバックまたはチャネル経由でProcess Managerに通知し、UIでのリアルタイム進捗表示を可能にする。

### 8. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス (PM側): `github.com/mattn/go-sqlite3` または標準 `database/sql`
- 依存性注入: `github.com/google/wire`
- 並行処理: Go標準 `sync`, `context`

## 関連ドキュメント
- [クラス図](./npc_persona_generator_class_diagram.md)
- [シーケンス図](./npc_persona_generator_sequence_diagram.md)
- [テスト設計](./npc_persona_generator_test_spec.md)
- [要件定義書](../requirements.md)
- [Term Translator Slice 仕様書](../term_translator/spec.md)
- [LLMクライアントインターフェース](../llm_client/llm_client_interface.md)
- [Config Store 仕様書](../config_store/spec.md)
