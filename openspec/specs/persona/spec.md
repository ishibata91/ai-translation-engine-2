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
- **AND** リクエスト生成前に `persona_text` 以外の NPC 基本情報と会話行をDBへ保存する
- **AND** 構築されたプロンプトの配列 `[]llm_client.Request` を返す（自身ではLLMクライアントを呼び出さない）
- **AND** `specs/log-guide.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: ペルソナ生成結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから性格・口調・背景セクションを抽出・パースする
- **AND** パースに成功したデータを `source_plugin + speaker_id` を一意キーとして `npc_personas` テーブルに対して UPSERT する
- **AND** `specs/log-guide.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. 独立性: ペルソナ生成データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: 外部のデータ構造を直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（例: `PersonaGenInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

本 slice は、自前の入力 DTO と保存 DTO を契約として公開しなければならない。persona slice は runtime 制御に依存してはならず、必要な外部資源は gateway 契約経由で利用しなければならない。

#### Scenario: 独自定義DTOによる初期化と生成処理の開始
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`PersonaGenInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いてペルソナ生成処理を完結できること

#### Scenario: persona slice は runtime を参照しない
- **WHEN** persona slice の contract と実装を参照する
- **THEN** queue、progress、resume など runtime 制御へ依存してはならない

#### Scenario: persona slice は gateway 契約を利用できる
- **WHEN** persona slice が永続化や LLM 依頼準備に必要な外部資源を使う
- **THEN** gateway 契約を通じて利用しなければならない

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
ペルソナDBは **`db/persona.db`** を使用し、`npc_personas` テーブルへUPSERTする。

### 7. ペルソナDBスキーマ

#### テーブル: `npc_personas`
| カラム             | 型                                | 説明                            |
| :----------------- | :-------------------------------- | :------------------------------ |
| `id`               | INTEGER PRIMARY KEY               | 自動採番ID                      |
| `speaker_id`       | TEXT                              | NPC識別子（SpeakerID）          |
| `editor_id`        | TEXT                              | NPCのEditor ID                  |
| `npc_name`         | TEXT                              | NPC名                           |
| `race`             | TEXT                              | 種族                            |
| `sex`              | TEXT                              | 性別                            |
| `voice_type`       | TEXT                              | 声の種類                        |
| `persona_text`     | TEXT NOT NULL                     | 生成されたペルソナテキスト      |
| `generation_request` | TEXT NOT NULL                   | LLMへ送信した生成リクエスト本文 |
| `status`           | TEXT NOT NULL                     | 状態 (`draft` / `generated`)    |
| `source_plugin`    | TEXT NOT NULL                     | ソースプラグイン名              |
| `updated_at`       | DATETIME                          | 更新日時                        |

`npc_personas` は `UNIQUE(source_plugin, speaker_id)` を持ち、同一 `speaker_id` が別プラグインに存在しても衝突してはならない。

#### テーブル: `npc_dialogues`
| カラム               | 型                  | 説明                                   |
| :------------------- | :------------------ | :------------------------------------- |
| `id`                 | INTEGER PRIMARY KEY | 自動採番ID                             |
| `persona_id`         | INTEGER NOT NULL    | `npc_personas.id` 参照                 |
| `source_plugin`      | TEXT NOT NULL       | ソースプラグイン名                     |
| `speaker_id`         | TEXT NOT NULL       | NPC識別子（検索補助）                  |
| `editor_id`          | TEXT                | 会話レコードのEditor ID                |
| `record_type`        | TEXT                | 会話レコード種別                       |
| `source_text`        | TEXT NOT NULL       | 原文                                   |
| `quest_id`           | TEXT                | 関連Quest ID (nullable)                |
| `is_services_branch` | INTEGER NOT NULL    | services分岐フラグ (0/1)               |
| `dialogue_order`     | INTEGER NOT NULL    | 元データ上の並び順                     |
| `updated_at`         | DATETIME            | 更新日時                               |

### 8. Pass 2での参照
Pass 2（本文翻訳）において、`SpeakerID` をキーとしてペルソナを検索し、LLMプロンプトのコンテキストに含める。

### 9. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス: `github.com/mattn/go-sqlite3` または `modernc.org/sqlite`
- 依存性注入: `github.com/google/wire`

### Requirement: ペルソナ生成ジョブ提案（Phase 1）はタスク境界で実行されなければならない
ペルソナ生成のPhase 1（`PreparePrompts`）は、UI同期呼び出しではなく `task.Bridge` / `task.Manager` が管理するタスク境界で実行されなければならない。システムは開始から完了/失敗まで同一タスクIDで追跡可能でなければならず、処理中のログは `specs/log-guide.md` に準拠した構造化ログとして出力しなければならない。

#### Scenario: UI起点でPhase 1がタスクとして実行される
- **WHEN** `MasterPersona` からペルソナ生成開始要求が送信される
- **THEN** システムは `task.Bridge` を介して `PreparePrompts` を非同期実行する
- **AND** タスク状態を `Pending` -> `Running` -> `Completed` または `Failed` に遷移させる

#### Scenario: LLM統合前段でも同一タスクIDで追跡できる
- **WHEN** Phase 1のリクエスト生成処理が完了する
- **THEN** システムは将来の `pkg/llm` 連携に引き継げる単一タスクIDを保持する
- **AND** 完了時点でリクエスト件数サマリを `info` ログに記録する

### Requirement: 重要度スコアリングはprobeを使用せず大文字フレーズ率で評価しなければならない
会話データの重要度スコアリングは、性能負荷の高い `probe` 依存処理を使用してはならない。英語ダイアログについては大文字フレーズ出現率を感情/強調シグナルとして評価し、日本語ダイアログについては当該スコアリングを適用してはならない。

#### Scenario: 英語ダイアログでは大文字フレーズ率がスコア計算に反映される
- **WHEN** 入力ダイアログが英語として判定される
- **THEN** システムは大文字フレーズ出現率を重要度スコアの一部として計算する
- **AND** クエスト優先度など既存の他特徴量と合成して順位付けする

#### Scenario: 日本語ダイアログでは大文字フレーズ率スコアリングをスキップする
- **WHEN** 入力ダイアログが日本語として判定される
- **THEN** システムは大文字フレーズ出現率の計算を行わない
- **AND** スコア算出は日本語で有効な他特徴量のみで実行する

### Requirement: ペルソナ保存は再開時も冪等でなければならない
MasterPersona の保存フェーズは、再試行または再開が発生しても同一 NPC を `source_plugin + speaker_id` で一意に識別し、重複レコードを作成せず、`overwrite_existing` に応じて更新または保持として確定保存しなければならない。保存が成功した行は `npc_personas.status` を英語値 `generated` に更新し、リクエスト生成時の `draft` と区別できなければならない。

#### Scenario: 再開後の再保存で重複が作成されない
- **WHEN** 一部保存済み状態でタスクを再開する
- **THEN** システムは未保存分のみ新規反映し、既存 NPC レコードは `source_plugin + speaker_id` 単位で更新または保持として扱わなければならない

#### Scenario: 保存失敗 request だけ再試行される
- **WHEN** 保存フェーズで一部 request が失敗する
- **THEN** 次回再開では失敗 request のみ保存再試行し、成功済み request は再保存してはならない

#### Scenario: 上書き有効時は既存行が更新される
- **WHEN** `overwrite_existing=true` で同一 `source_plugin + speaker_id` の保存対象が存在する
- **THEN** システムは既存行を更新し、重複行を新規作成してはならない

#### Scenario: 上書き無効時は既存行を保持する
- **WHEN** `overwrite_existing=false` で同一 `source_plugin + speaker_id` の保存対象が存在する
- **THEN** システムは既存行を保持し、既存行の内容を変更してはならない

#### Scenario: 保存成功時に status は generated になる
- **WHEN** `SavePersona` がペルソナ本文の保存を完了する
- **THEN** システムは対象行の `status` を `generated` に更新しなければならない
- **AND** `persona_text` が確定した行を `draft` のまま残してはならない

### Requirement: リクエスト生成前の persona 下書き保存は原データ属性を正しく保持しなければならない
`PreparePrompts` 実行時に `npc_personas` / `npc_dialogues` へ事前保存する属性は、JSON抽出元の意味を変えずに保持しなければならない。`race` へ `record_type` 等の別属性を混入させてはならない。`source_plugin` が欠損している場合は入力ファイル名から `*.esm|*.esl|*.esp` を補完し、抽出不能時は `UNKNOWN` を設定しなければならない。加えて、`npc_personas.status` には英語値 `draft` を保存し、リクエスト生成済みだがペルソナ本文未保存の状態を明示しなければならない。

#### Scenario: npc_personas の属性が抽出元と一致する
- **WHEN** リクエスト生成前に `npc_personas` へ下書き保存する
- **THEN** `race` には NPC の種族値のみが保存されなければならない
- **AND** `sex` / `voice_type` / `source_plugin` は空でなく保存されなければならない（入力に値がある場合）

#### Scenario: source_plugin 欠損時はファイル名から補完される
- **WHEN** 入力データの `source_plugin` が空で、入力パスまたは入力名に `hoge.esm` / `hoge.esl` / `hoge.esp` が含まれる
- **THEN** システムは該当拡張子付きファイル名を `source_plugin` として保存しなければならない

#### Scenario: source_plugin を補完できない場合は UNKNOWN を設定する
- **WHEN** `source_plugin` が空で、入力情報から `*.esm|*.esl|*.esp` を抽出できない
- **THEN** システムは `source_plugin` に `UNKNOWN` を保存しなければならない

#### Scenario: npc_dialogues の editor_id はレスポンス欠損時にフォールバックされる
- **WHEN** dialogue response の `editor_id` が空で、dialogue group 側に `editor_id` がある
- **THEN** `npc_dialogues.editor_id` には group 側 `editor_id` を保存しなければならない

#### Scenario: npc_dialogues は原文中心で保存される
- **WHEN** ペルソナ用途の会話データを `npc_dialogues` に保存する
- **THEN** システムは `source_text` を保存しなければならない
- **AND** `translated_text` の更新を必須処理として扱ってはならない

#### Scenario: リクエスト生成時に status は draft になる
- **WHEN** `PreparePrompts` が `npc_personas` へ下書き保存と生成リクエスト保存を完了する
- **THEN** システムは対象行の `status` に `draft` を保存しなければならない
- **AND** `generation_request` が存在しても `status` を `generated` にしてはならない

### Requirement: MasterPersona はユーザープロンプト入力カードと読み取り専用システムプロンプトカードを表示しなければならない
`MasterPersona` 画面は、ユーザーが編集できるユーザープロンプト入力カードと、送信時に適用されるシステムプロンプトを読み取り専用で確認できるカードを同一画面内に表示しなければならない。

#### Scenario: ユーザープロンプトを画面上で編集できる
- **WHEN** ユーザーが MasterPersona 画面を開く
- **THEN** システムはユーザープロンプト入力欄を表示し、ユーザーが内容を編集できなければならない

#### Scenario: システムプロンプトは読み取り専用で表示される
- **WHEN** ユーザーが MasterPersona 画面で prompt 設定を確認する
- **THEN** システムはシステムプロンプト欄を表示しなければならない
- **AND** 当該欄は読み取り専用であり、ユーザー編集を受け付けてはならない

### Requirement: MasterPersona の固定補完値はシステムプロンプトへ分離されなければならない
MasterPersona の prompt 構築時にシステムが自動注入する固定説明、制約、出力形式、入力データの補足は、ユーザープロンプトではなくシステムプロンプトとして扱われなければならない。ユーザープロンプトにはユーザーが意図的に編集する可変指示のみを保持しなければならない。

#### Scenario: 自動補完値はユーザープロンプトから除外される
- **WHEN** MasterPersona が既定 prompt を初期化または再構築する
- **THEN** システムが差し込む固定文言はユーザープロンプトに含めてはならない
- **AND** 同じ固定文言はシステムプロンプトとして保持されなければならない

#### Scenario: 画面表示と実際の送信 prompt の責務が一致する
- **WHEN** MasterPersona が LLM へ送信する prompt を組み立てる
- **THEN** ユーザープロンプトカードの内容は user prompt として使用されなければならない
- **AND** 読み取り専用カードに表示された内容は system prompt として使用されなければならない

### Requirement: MasterPersona 一覧は保存済みダイアログ件数を表示しなければならない
MasterPersona のペルソナ一覧は、`npc_personas.dialogue_count` のような生成時スナップショット値ではなく、現在 `npc_dialogues` に保存されている関連会話件数を表示しなければならない。システムは一覧取得時に関連ダイアログを集計して表示用 DTO に反映し、同じ DTO に `npc_personas.status` の英語値 `draft` / `generated` を含めなければならない。フロントエンドはこの状態値を使って `下書き` / `生成済み` として表示しなければならない。

#### Scenario: 一覧件数は関連ダイアログ数から算出される
- **WHEN** ユーザーが MasterPersona のペルソナ一覧を開く
- **THEN** システムは各ペルソナについて `npc_dialogues` の関連件数を集計して返さなければならない
- **AND** `npc_personas.dialogue_count` を一覧表示の根拠として用いてはならない

#### Scenario: ダイアログ件数が更新されると一覧表示も追従する
- **WHEN** 既存ペルソナに紐づく `npc_dialogues` が追加または削除される
- **THEN** 次回一覧取得時のセリフ数は最新の関連件数を表示しなければならない

#### Scenario: 一覧は status に応じた表示名を出す
- **WHEN** 一覧取得結果の `status` が `draft` または `generated` を含む
- **THEN** フロントエンドは `draft` を `下書き`、`generated` を `生成済み` として表示しなければならない
- **AND** 全件を固定の完了表示にしてはならない

### Requirement: MasterPersona 一覧はステータスでフィルタできなければならない
MasterPersona 一覧は、既存の検索語とプラグイン絞り込みに加えて、`draft` / `generated` の状態で絞り込みできなければならない。フィルタ条件は一覧表示にのみ適用され、元の一覧データを破壊してはならない。

#### Scenario: 下書きだけを絞り込める
- **WHEN** ユーザーがステータスフィルタで `下書き` を選択する
- **THEN** 一覧には `status='draft'` のペルソナだけが表示されなければならない

#### Scenario: 生成済みだけを絞り込める
- **WHEN** ユーザーがステータスフィルタで `生成済み` を選択する
- **THEN** 一覧には `status='generated'` のペルソナだけが表示されなければならない

#### Scenario: ステータス解除で全件に戻せる
- **WHEN** ユーザーがステータスフィルタを未選択に戻す
- **THEN** システムは他の検索条件だけを維持しつつ、全ステータスの一覧を再表示しなければならない

### Requirement: Persona 詳細は生成リクエストを確認できなければならない
Persona 詳細表示は、従来の `RAW response` ではなく、実際に LLM へ送信した生成リクエストを `生成リクエスト` として返却・表示しなければならない。システムは詳細取得用データに当該 request 本文を保持し、ユーザーが生成条件を検証できるようにしなければならない。

#### Scenario: 詳細画面に生成リクエストが表示される
- **WHEN** ユーザーがペルソナ詳細を開く
- **THEN** システムは `生成リクエスト` 欄に実際に送信した request 本文を表示しなければならない
- **AND** `RAW response` を同名の検証欄として表示してはならない

#### Scenario: 完了後も生成リクエストを再確認できる
- **WHEN** タスク完了後にアプリを再表示し、既存ペルソナ詳細を開く
- **THEN** システムは queue の一時データに依存せず、保存済みの生成リクエストを返却しなければならない

## 関連ドキュメント
- [テスト設計](./npc_personaerator_test_spec.md)
- [要件定義書](../requirements.md)
- [Terminology Slice 仕様書](../terminology/spec.md)
- [LLMクライアントインターフェース](../llm_client/llm_client_interface.md)
- [Config 仕様書](../config/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `standard_test_spec.md` と `log-guide.md` に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（`standard_test_spec.md` 参照）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（`log-guide.md` 参照）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（`log-guide.md` 参照）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（`log-guide.md` 参照）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（`log-guide.md` の AI デバッグ運用参照）でログと仕様書をAIに渡し修正させる。
