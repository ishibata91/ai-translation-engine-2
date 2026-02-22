# 用語翻訳・Mod用語DB保存 (Term Translator Slice) 仕様書

## 概要
Modデータから抽出された固有名詞（NPC名、アイテム名、場所名等）をLLMで翻訳し、その結果をMod専用の用語SQLiteデータベースへ保存する機能である。
本機能は2-Pass Systemにおける **Pass 1: 用語翻訳 (Term Translation)** の中核を担い、Pass 2（本文翻訳）で参照される用語辞書を構築する。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則（データ構造やDB操作の共通化）を捨て、**本Slice自身が「Mod用語テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- 現行Python版では `context_builder.py` の `build_term_requests` が用語翻訳リクエストを生成し、`translator.py` の `translate_batch` がLLM翻訳を実行し、`term_cache.py` の `create_term_cache` が結果をSQLiteへ保存している。
- これら3つの責務が複数モジュールに分散しており、Go版では単一のVertical Sliceとして凝集させる。
- 構造化データの作成（会話ツリー解析等）は本Sliceの責務外であり、別Sliceが担当する。

## スコープ
### 本Sliceが担う責務
1. **用語翻訳リクエストの生成**: ロード済みの `ExtractedData` から、用語翻訳対象のレコード（名詞類）を抽出し、翻訳リクエストを構築する。
2. **既存辞書からの参照用語検索**: Dictionary Builder Sliceが構築した辞書DB（公式DLC辞書）から、翻訳対象テキストに関連する既訳用語を検索し、LLMプロンプトのコンテキストとして提供する。
3. **LLMによる用語翻訳の実行**: 構築した翻訳リクエストをLLMに送信し、翻訳結果を取得する。
4. **Mod用語DBへの保存**: 翻訳結果をMod専用のSQLiteデータベースに永続化し、Pass 2で参照可能にする。

### 本Sliceの責務外
- xTranslator XMLからの辞書DB構築（Dictionary Builder Sliceの責務）
- 会話ツリー解析・文脈構築（Context Engine Sliceの責務）
- Pass 2の本文翻訳（別Sliceの責務）

## 要件

### 独立性: 用語翻訳対象データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライス(LoaderSlice等)のDTOへの依存を排除するため。
**Migration**: 外部の `ExtractedData` などを直接参照する方式から、本スライス独自のパッケージ内に入力用DTO（例: `TermTranslatorInput`）を定義し、それを受け取るインターフェースへ移行する。マッピングは呼び出し元（オーケストレーター層）の責務とする。

#### Scenario: 独自定義DTOによる初期化と翻訳処理の開始
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`TermTranslatorInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、提供された内部データ構造のみを用いて用語翻訳リクエストを生成できること

### 1. 用語翻訳対象レコードの定義
本Sliceは、`ExtractedData` に含まれる以下のドメインモデルから用語翻訳リクエストを生成する。

| ドメインモデル | 対象レコードタイプ                                                                                     | 説明                                                 |
| :------------- | :----------------------------------------------------------------------------------------------------- | :--------------------------------------------------- |
| `NPC`          | `NPC_:FULL`, `NPC_:SHRT`                                                                               | NPC名、NPCの短い名前                                 |
| `Item`         | `WEAP:FULL`, `ARMO:FULL`, `AMMO:FULL`, `MISC:FULL`, `KEYM:FULL`, `ALCH:FULL`, `BOOK:FULL`, `INGR:FULL` | 武器・防具・弾薬・雑貨・鍵・錬金術品・本・素材の名称 |
| `Magic`        | `SPEL:FULL`, `MGEF:FULL`, `ENCH:FULL`                                                                  | 魔法・魔法効果・エンチャントの名称                   |
| `Message`      | `MESG:FULL`                                                                                            | メッセージ名                                         |
| `Location`     | `LCTN:FULL`, `CELL:FULL`, `WRLD:FULL`                                                                  | ロケーション・セル・ワールドスペース名               |
| `Quest`        | `QUST:FULL`                                                                                            | クエスト名                                           |

**共通設定（Config）化による再利用**: 上記の「用語翻訳対象のレコードタイプ定義」は、本Sliceの内部にハードコードするのではなく、**システム共通のConfig（設定情報定義）として切り出して定義し、DI等で注入**する。これにより、Dictionary Builder Sliceの抽出対象REC定義と同一の設計思想を共有し、将来的な対象拡張を容易にする。

#### 1.1 NPC名のペア翻訳（FULL + SHRT 同時翻訳）
NPCは `NPC_:FULL`（フルネーム、例: `"Jon Battle-Born"`）と `NPC_:SHRT`（短い名前、例: `"Jon"`）の2つのレコードタイプを持つ。これらは**必ずペアとして同一のLLMリクエストで同時に翻訳**する。

**ペア翻訳の理由**:
- FULLとSHRTは同一NPCの名前の異なる表現であり、翻訳の一貫性を保つために同一コンテキストで翻訳する必要がある。
- 例: FULL `"Jon Battle-Born"` → `"ジョン・バトルボーン"` に対し、SHRT `"Jon"` → `"ジョン"` と整合させる。別々に翻訳すると、SHRTがFULLと矛盾する訳になるリスクがある。

**リクエスト構築ルール**:
1. `TermRequestBuilder` は、同一EditorIDを持つ `NPC_:FULL` と `NPC_:SHRT` を**1つの `TermTranslationRequest` にペアリング**する。
2. ペアリングされたリクエストの `SourceText` にはFULLのテキストを設定し、`ShortName` フィールドにSHRTのテキストを設定する。
3. LLMプロンプトでは、FULLとSHRTの両方を提示し、それぞれの翻訳を同時に要求する。
4. SHRTが存在しないNPC（FULLのみ）は、通常の単独リクエストとして処理する。
5. 翻訳結果は `TermTranslationResult` としてFULL・SHRTそれぞれ個別のレコードに分解してMod用語DBに保存する。

### 2. 既存辞書からの参照用語検索
- 翻訳対象テキストからキーワードを抽出し、Dictionary Builder Sliceが構築した辞書DB（公式DLC辞書）および追加辞書DBから関連用語を検索する。
- 検索結果はLLMプロンプトの `reference_terms`（参照用語リスト）としてコンテキストに含める。
- 辞書DBへの接続は `*sql.DB` をDIで受け取り、検索ロジックは本Slice内にカプセル化する。

#### 2.1 検索戦略: 貪欲部分一致（Greedy Partial Match）
本Sliceの辞書検索は、**完全一致（Exact Match）を最優先**しつつ、**長さベースの貪欲マッチング（Greedy Longest Match）** を適用して参照用語を決定する。曖昧検索（FTS5 MATCH等）はハルシネーション防止のため補助的にのみ使用する。

**検索フロー（優先順）**:
1. **ソーステキスト全文の完全一致**: 翻訳対象テキスト全体を辞書の `source` カラムと照合する。
2. **キーワード完全一致**: ソーステキストから抽出したキーワード群を辞書の `source` カラムと `IN` 句で照合する。
3. **NPC名の貪欲部分一致**（後述 §2.2）。
4. **貪欲最長一致フィルタリング**: 上記で得られた全候補に対し、以下のアルゴリズムで重複を排除する。
   - 候補を**文字数降順**でソートする。
   - 長い候補から順にソーステキスト内での出現位置を確認する。
   - 既にマッチ済みの文字区間と**重複しない**場合のみ採用する。
   - これにより、`"Broken Tower Redoubt"` が辞書にある場合、部分文字列 `"Broken"` の単独マッチは抑制される。

#### 2.2 NPC名の貪欲部分一致
NPC名（`NPC_:FULL`, `NPC_:SHRT`）は、他のレコードタイプと異なり**部分一致検索**を積極的に行う。これは以下のユースケースを解決するためである。

- **家族名（苗字）の解決**: Modで追加されたNPCの苗字（例: `"Battle-Born"`）が会話テキストに単独で出現する場合、フルネーム `"Jon Battle-Born"` の辞書エントリから苗字部分の訳を参照できるようにする。
- **名詞のみの参照**: NPC名レコード専用のFTS5テーブル（`npc_terms_fts`）を用いて、キーワードに対する部分一致検索を行う。

**NPC部分一致の検索ルール**:
1. ソーステキストから抽出したキーワードのうち、完全一致検索で**消費されなかった**キーワードを対象とする。
2. `npc_terms_fts` テーブルに対してFTS5 MATCH検索を実行する。
3. 結果は貪欲最長一致フィルタリングの**対象外**とし、フルネームとファーストネーム/苗字の両方をLLMコンテキストに提供する。
4. NPC名レコード自体の翻訳時（`is_npc=true`）は、全キーワードを部分一致検索の対象とする。

#### 2.3 英語形態変化への対応（ステミング）
英語の用語は複数形（`Swords`）、所有格（`Auriel's`）などの形態変化を持つ。辞書検索時にこれらの変化形を原形に正規化することで、辞書ヒット率を向上させる。

**ステミングライブラリ**: Snowball Stemmer（Go実装: `github.com/kljensen/snowball`）を使用する。

**適用箇所**:
1. **キーワード抽出時**: ソーステキストからキーワードを抽出する際、各キーワードに対してSnowball English Stemmerを適用し、ステム（語幹）を取得する。
2. **辞書検索時（キーワード完全一致）**: 原形キーワードでの完全一致検索がヒットしなかった場合、ステム化したキーワードで辞書の `source` カラムに対して再検索する。辞書側の `source` もステム化して比較する。
3. **NPC部分一致検索時**: FTS5 MATCH検索のクエリにもステム化を適用する。

**ステミングルール**:
- 所有格の `'s` は、ステミング前に除去する（例: `"Auriel's"` → `"Auriel"` → ステム化）。
- ステム化の結果が元のキーワードと同一の場合、重複検索は行わない。
- ステム化による検索結果は、原形キーワードによる完全一致結果より**低優先**とする。
- 強制翻訳（§6）の判定には**ステム化を適用しない**（原文全文の厳密な完全一致のみ）。

**具体例**:
| ソーステキスト     | キーワード            | ステム   | 辞書エントリ              | マッチ |
| :----------------- | :-------------------- | :------- | :------------------------ | :----- |
| `"Daedric Swords"` | `Swords`              | `sword`  | `Sword` → stem: `sword`   | ✓      |
| `"Auriel's Bow"`   | `Auriel's` → `Auriel` | `auriel` | `Auriel` → stem: `auriel` | ✓      |
| `"Dragon Priests"` | `Priests`             | `priest` | `Priest` → stem: `priest` | ✓      |

### 3. LLM翻訳の実行
- 用語翻訳リクエストごとに、レコードタイプに応じた最適なシステムプロンプトを動的に生成する。
- LLMクライアントインターフェース（`infrastructure/llm_client`）を通じて翻訳を実行する。
- リトライ（指数バックオフ）とタイムアウト制御を備える。
- 並列翻訳（Goroutine）により処理を高速化する。
- 翻訳対象テキストが既に日本語を含む場合はスキップする。

### 4. Mod用語DBへの保存（カプセル化された永続化）
- プロセスマネージャーから `*sql.DB` などの**「DBのプーリング・接続管理のためだけのインフラモジュール」**のみをDIで受け取る。
- 本Slice内の `ModTermStore` がMod用語テーブルに対するすべての操作（テーブル生成・INSERT/UPSERT・FTS5インデックス管理）を単独で完結させる。
- Mod用語DBは**Modごとに独立したSQLiteファイル**として生成し、Pass 2で `additional_db_paths` として参照される。

### 5. Mod用語DBスキーマ

#### テーブル: `mod_terms`
| カラム          | 型                                | 説明                              |
| :-------------- | :-------------------------------- | :-------------------------------- |
| `id`            | INTEGER PRIMARY KEY AUTOINCREMENT | 自動採番ID                        |
| `original_en`   | TEXT NOT NULL                     | 原文（英語、小文字正規化）        |
| `translated_ja` | TEXT NOT NULL                     | 翻訳結果（日本語）                |
| `record_type`   | TEXT NOT NULL                     | レコードタイプ（例: `NPC_ FULL`） |
| `editor_id`     | TEXT                              | Editor ID（例: `DLC1AurielsBow`） |
| `source_plugin` | TEXT                              | ソースプラグイン名                |
| `created_at`    | DATETIME                          | 作成日時                          |
| UNIQUE          |                                   | `(original_en, record_type)`      |

#### 仮想テーブル: `mod_terms_fts` (FTS5)
| カラム          | 説明                   |
| :-------------- | :--------------------- |
| `original_en`   | 原文（全文検索用）     |
| `translated_ja` | 翻訳結果（全文検索用） |

- `content='mod_terms'`, `content_rowid='id'`, `tokenize='unicode61'`
- INSERT/UPDATE/DELETEトリガーで `mod_terms` テーブルと自動同期する。

#### 仮想テーブル: `npc_terms_fts` (FTS5)
- NPC名レコード（`NPC_ FULL`, `NPC_ SHRT`）専用のFTS5テーブル。部分一致検索の対象とする。
- スキーマは `mod_terms_fts` と同一。

### 6. 強制翻訳（Forced Translation）
- 辞書DBに**ソーステキスト全文の完全一致**する既訳が存在する場合、LLMを呼び出さずに辞書の訳をそのまま採用する（Forced Translation）。
- 貪欲部分一致で得られた参照用語は強制翻訳の対象とはならず、あくまでLLMプロンプトのコンテキストとして使用する。
- これによりAPI呼び出しコストを削減し、公式訳との一貫性を保証する。

### 7. 進捗通知
- 翻訳の進捗（完了数/総数）をコールバックまたはチャネル経由でProcess Managerに通知し、UIでのリアルタイム進捗表示を可能にする。

### 8. ライブラリの選定
- LLMクライアント: `infrastructure/llm_client` インターフェース（プロジェクト共通）
- DBアクセス (PM側): `github.com/mattn/go-sqlite3` または標準 `database/sql`
- 依存性注入: `github.com/google/wire`
- 並行処理: Go標準 `sync`, `context`
- ステミング: `github.com/kljensen/snowball`（Snowball English Stemmer）

## 関連ドキュメント
- [クラス図](./term_translator_class_diagram.md)
- [シーケンス図](./term_translator_sequence_diagram.md)
- [テスト設計](./term_translator_test_spec.md)
- [Dictionary Builder Slice 仕様書](../dictionary_builder/spec.md)
- [LLMクライアントインターフェース](../llm_client/llm_client_interface.md)
- [Config Store 仕様書](../config_store/spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
