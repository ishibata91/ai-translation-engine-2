# NPCペルソナ生成 (NPC Personaerator Slice) 仕様書

## 概要
NPCごとの会話データ（最大100件）を収集し、LLMにリクエストしてその NPC の性格・口調・背景を要約した「ペルソナ」を自動生成する機能である。
本機能は2-Pass Systemにおける **Pass 1** の一部として実行され、生成されたペルソナはPass 2（本文翻訳）で話者プロファイルとして参照される。

当機能は、バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト（ジョブ）生成」と「結果保存」の2段階に分離する **2フェーズモデル（提案/保存モデル）** を採用する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「ペルソナテーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `context_builder.py` が話者の種族・声タイプから口調を推定しているが、NPCの個別の性格や背景は反映されていない。
- 会話データからLLMでペルソナを生成することで、翻訳時の口調一貫性を大幅に向上させる。
- **2フェーズモデルへの移行**: LLMへの通信制御をスライスから分離し、Job Queue / Pipeline へ委譲することで、ネットワーク分断やバッチAPI待機に対する堅牢性を確保する。
- コンテキスト長の制約があるため、ペルソナ生成前にトークン利用量を事前計算し、超過を防止する必要がある。

## スコープ
### 本Sliceが担う責務
1. **ペルソナ生成ジョブの生成 (Phase 1: Propose)**: NPC会話データの収集、重要度スコアリング、トークン事前計算、コンテキスト長評価を行い、LLMリクエスト群（ジョブ）を構築して返す。
2. **ペルソナの保存 (Phase 2: Save)**: LLMからのレスポンス群を受け取り、パース・バリデーションを行った上でMod用語DBまたは専用テーブルに保存する。
3. **ペルソナの取得**: SpeakerIDに基づき、保存済みのペルソナテキストを取得する。

### 本Sliceの責務外
- LLMへの実際のHTTP通信制御（Job Queue / LLM Client の責務）
- 会話ツリーの構造解析（Lore Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）
- 用語翻訳（Terminology Sliceの責務）

## 要件

### 1. 2フェーズモデル（提案/保存モデル） (Propose/Save Model)
**Reason**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成」と「結果保存」の2段階に分離し、通信制御をインフラ層（JobQueue/Pipeline）へ委譲する。

#### Scenario: ペルソナ生成ジョブの提案 (Phase 1: Propose)
- **WHEN** プロセスマネージャーから `PersonaGenInput` 形式のデータを受け取った
- **THEN** 各NPCに対して会話データの収集・選別、コンテキスト長評価を行う
- **AND** 構築されたプロンプトの配列 `[]llm_client.Request` を返す（自身ではLLMクライアントを呼び出さない）
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: ペルソナ生成結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから性格・口調・背景セクションを抽出・パースする
- **AND** パースに成功したデータを `npc_personas` テーブルに対して UPSERT する
- **AND** `specs/architecture.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: ペルソナ生成データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（例: `PersonaGenInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化と生成処理の開始
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`PersonaGenInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いてペルソナ生成処理を完結できること

### 3. NPC会話データの収集

本Sliceは、`PersonaGenInput` から、NPCごとの会話データを収集する。

**収集ルール**:
1. 各NPCにつき最大 **100件** の会話データを収集する。
2. 100件を超える場合は、**重要度スコアリング**（後述 §3.1）によりスコアの高い上位100件を選別する。

#### 3.1 重要度スコアリング (Importance Scoring)
NPCの会話データが100件を超える場合、各会話ペア（英語原文・日本語訳）に対して**重要度スコア**を算出する。

**スコアリング要素**:
1. **固有名詞スコア (Proper Noun Score)**: 翻訳済みの固有名詞のマッチ件数。
2. **感情スコア (Emotion Score)**: 句読点、全大文字、感情語の出現回数。
3. **カテゴリ優先度**: クエスト関連会話を優遇。

### 4. トークン利用量の事前計算
ペルソナ生成リクエストの送信前に、NPCごとの想定トークン利用量を計算し、LLMのコンテキストウィンドウ超過を防止する。
超過が見込まれる場合は、会話データを優先度順に削減し、最低 **10件** を確保した上で上限内に収める。

### 5. LLMによるペルソナ生成
収集した会話データとNPC属性を元に、LLMリクエストを構築する。

**出力フォーマット**:
```
性格特性: <NPCの性格を1-2文で要約>
話し方の癖: <口調・語尾・特徴的な表現を1-2文で要約>
背景設定: <NPCの立場・役割・背景を1-2文で要約>
```

### 6. ペルソナの永続化
生成されたペルソナをSQLiteに保存し、Pass 2で参照可能にする。
ペルソナDBは **Mod用語DBと同一のSQLiteファイル** 内に専用テーブルとして格納する。

### 7. ペルソナDBスキーマ

#### テーブル: `npc_personas`
| カラム             | 型                                | 説明                            |
| :----------------- | :-------------------------------- | :------------------------------ |
| `id`               | INTEGER PRIMARY KEY AUTOINCREMENT | 自動採番ID                      |
| `speaker_id`       | TEXT NOT NULL UNIQUE              | NPC識別子（SpeakerID）          |
| `editor_id`        | TEXT                              | NPCのEditor ID                  |
| `npc_name`         | TEXT                              | NPC名                           |
| `race`             | TEXT                              | 種族                            |
| `sex`              | TEXT                              | 性別                            |
| `voice_type`       | TEXT                              | 声の種類                        |
| `persona_text`     | TEXT NOT NULL                     | 生成されたペルソナテキスト      |
| `dialogue_count`   | INTEGER NOT NULL                  | ペルソナ生成に使用した会話件数  |
| `estimated_tokens` | INTEGER                           | 推定トークン利用量              |
| `source_plugin`    | TEXT                              | ソースプラグイン名              |
| `created_at`       | DATETIME                          | 作成日時                        |
| `updated_at`       | DATETIME                          | 更新日時                        |

### 8. Pass 2での参照
Pass 2（本文翻訳）において、`SpeakerID` をキーとしてペルソナを検索し、LLMプロンプトのコンテキストに含める。

### 9. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス: `github.com/mattn/go-sqlite3` または `modernc.org/sqlite`
- 依存性注入: `github.com/google/wire`

## 関連ドキュメント
- [クラス図](./npc_personaerator_class_diagram.md)
- [シーケンス図](./npc_personaerator_sequence_diagram.md)
- [テスト設計](./npc_personaerator_test_spec.md)
- [要件定義書](../requirements.md)
- [Terminology Slice 仕様書](../terminology/spec.md)
- [LLMクライアントインターフェース](../llm_client/llm_client_interface.md)
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
