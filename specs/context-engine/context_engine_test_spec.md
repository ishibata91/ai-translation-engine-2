# コンテキストエンジン テスト設計 (Context Engine Test Spec)

本設計は `refactoring_strategy.md` セクション 6（テスト戦略）および セクション 7（構造化ログ基盤）に厳密に準拠し、個別関数の細粒度なユニットテストを作成せず、Vertical Slice の Contract に対する網羅的なパラメタライズドテスト（Table-Driven Test）を定義する。

## 1. テスト方針

1. **細粒度ユニットテストの排除**: 内部処理や個別関数単位の振る舞いテストは作成しない。
2. **網羅的パラメタライズドテスト**: すべてのテストを Table-Driven Test として実装し、インメモリ SQLite (`:memory:`) や適切なモック（ToneResolver, PersonaLookup, TermLookup, SummaryLookup）を用いて Contract 全体の振る舞いを検証する。
3. **構造化ログの強制検証**: テスト実行時においても、必ず `context.Context` （独自の TraceID を内包）を引き回し、slog を用いた JSON 形式での出力を含める。

---

## 2. パラメタライズドテストケース (Table-Driven Tests)

各Contractに対する入力（初期状態 + アクション）と期待されるアウトプット（戻り値 + 変更後の状態）を表として定義し、ループ内で検証する。

### 2.1 ToneResolver 統合テスト
種族、ボイスタイプ、性別に応じた適切な口調指示の解決プロセスを検証する。

| ケースID | 目的                       | 初期状態 (入力/DB)                        | アクション (関数呼出)                      | 期待される結果 (出力 / 状態)                 |
| :------- | :------------------------- | :---------------------------------------- | :----------------------------------------- | :------------------------------------------- |
| TONE-01  | 種族マッピング適用         | configにKhajiitRaceのマッピング有         | `Resolve("KhajiitRace", "", "Male")`       | カジート固有の口調指示が含まれる             |
| TONE-02  | ボイスタイプマッピング適用 | configにMaleCommanderのマッピング有       | `Resolve("", "MaleCommander", "Male")`     | MaleCommander固有の口調指示が含まれる        |
| TONE-03  | 種族・ボイスタイプの複合   | configにOrcRace, MaleCowardのマッピング有 | `Resolve("OrcRace", "MaleCoward", "Male")` | 種族指示とボイスタイプ指示が改行で結合される |
| TONE-04  | ボイスタイプ不明(女性)     | 該当マッピング無                          | `Resolve("", "", "Female")`                | `"標準的な女性の口調"` を含む指示が返る      |
| TONE-05  | ボイスタイプ不明(男性)     | 該当マッピング無                          | `Resolve("", "", "Male")`                  | `"標準的な男性の口調"` を含む指示が返る      |
| TONE-06  | 全属性空(フォールバック)   | 該当マッピング無                          | `Resolve("", "", "")`                      | デフォルトの男性口調が返る。パニックしない   |

### 2.2 PersonaLookup 統合テスト (SQLite)
NPCペルソナの取得処理を検証する。

| ケースID | 目的                   | 初期状態 (入力/DB)                           | アクション (関数呼出)            | 期待される結果 (出力 / 状態)   |
| :------- | :--------------------- | :------------------------------------------- | :------------------------------- | :----------------------------- |
| PERS-01  | 存在するペルソナ取得   | `npc_personas`にSpeakerID="NPC001"が存在する | `FindBySpeakerID("NPC001")`      | 保存した `persona_text` が返る |
| PERS-02  | 存在しないペルソナ取得 | `npc_personas`に"NONEXISTENT"が存在しない    | `FindBySpeakerID("NONEXISTENT")` | `nil, nil`が返る。エラーなし   |
| PERS-03  | 空文字のID指定         | -                                            | `FindBySpeakerID("")`            | `nil, nil`が返る。エラーなし   |

### 2.3 TermLookup 統合テスト (SQLite)
辞書DB・Mod用語DBからの複数条件による用語参照機構を検証する。

| ケースID | 目的                           | 初期状態 (入力/DB)                                              | アクション (関数呼出)                      | 期待される結果 (出力 / 状態)                                 |
| :------- | :----------------------------- | :-------------------------------------------------------------- | :----------------------------------------- | :----------------------------------------------------------- |
| TERM-01  | 完全一致検索                   | 辞書に "Iron Sword"→"鉄の剣" を登録                             | `Search("Iron Sword")`                     | `ForcedTranslation`="鉄の剣", `ReferenceTerms`に"Iron Sword" |
| TERM-02  | キーワード完全一致検索         | 辞書に "Iron"→"鉄", "Sword"→"剣" を登録                         | `Search("Iron Sword of Fire")`             | `ForcedTranslation`=nil, `ReferenceTerms`に"Iron","Sword"    |
| TERM-03  | 貪欲最長一致による短い候補抑制 | 辞書に "Broken Tower Redoubt"→"壊...", "Broken"→"壊れた" を登録 | `Search("Broken Tower Redoubt is nearby")` | `ReferenceTerms`に"Broken Tower Redoubt"のみ含まれる         |
| TERM-04  | NPC名の部分一致検索            | NPC FTS5テーブルに "Jon Battle-Born" を登録                     | `Search("Battle-Born family")`             | `ReferenceTerms`に"Jon Battle-Born"が含まれる                |
| TERM-05  | ステミング対応検索             | 辞書に "Sword"→"剣" を登録                                      | `Search("Daedric Swords")`                 | `ReferenceTerms`に"Sword"が含まれる                          |
| TERM-06  | 所有格の除去と検索             | 辞書に "Auriel"→"アウリエル" を登録                             | `Search("Auriel's Bow")`                   | `ReferenceTerms`に"Auriel"が含まれる                         |
| TERM-07  | 空の辞書DB検索                 | DBにデータが登録されていない空のDB                              | `Search("anything")`                       | 空のリストとnilが返る。エラーなし                            |
| TERM-08  | 複数DB横断検索                 | DB1: "Iron"→"鉄", DB2: "Steel"→"鋼"                             | `Search("Iron and Steel")`                 | `ReferenceTerms`に"Iron"と"Steel"が含まれる                  |

### 2.4 SummaryLookup 統合テスト (SQLite)
会話・クエスト要約の取得処理を検証する。

| ケースID | 目的             | 初期状態 (入力/DB)                     | アクション (関数呼出)                | 期待される結果 (出力 / 状態)   |
| :------- | :--------------- | :------------------------------------- | :----------------------------------- | :----------------------------- |
| SUMM-01  | 会話要約取得     | `summaries`に`DIAL001`(`dialogue`)登録 | `FindDialogueSummary("DIAL001")`     | 保存した `summary_text` が返る |
| SUMM-02  | クエスト要約取得 | `summaries`に`QUST001`(`quest`)登録    | `FindQuestSummary("QUST001")`        | 保存した `summary_text` が返る |
| SUMM-03  | 存在しないID検索 | 対象レコードなし                       | `FindDialogueSummary("NONEXISTENT")` | `nil, nil`が返る。エラーなし   |
| SUMM-04  | 種別不一致検索   | `summaries`に`DIAL001`(`dialogue`)登録 | `FindQuestSummary("DIAL001")`        | `nil, nil`が返る。エラーなし   |

### 2.5 ContextEngine 統合テスト（エンドツーエンド）
抽出データ (ExtractedData) 全体から各リクエストモデル (TranslationRequest) への構築フローを一気通貫で検証する（Lookup系はモックまたはインメモリDBにて対応）。

| ケースID | 目的                           | 初期状態 (入力/DB)                              | アクション (関数呼出)                    | 期待される結果 (出力 / 状態)                                 |
| :------- | :----------------------------- | :---------------------------------------------- | :--------------------------------------- | :----------------------------------------------------------- |
| CENG-01  | INFO NAM1リクエスト生成        | 1件のDialogueGroup(DialogueResponse1件+NPC情報) | `BuildTranslationRequests(data, config)` | INFO NAM1のリクエストが1件生成され、Speaker情報が付与される  |
| CENG-02  | 前回発言のトラッキング         | PlayerText="Hello", Response 2件                | `BuildTranslationRequests(data, config)` | リク1のPrevious="Hello", リク2のPrevious=(リク1のText)       |
| CENG-03  | DIAL FULLリクエスト生成        | PlayerText="Choose wisely"                      | `BuildTranslationRequests(data, config)` | DIAL FULLのリクエスト生成、TopicNameが算出・設定される       |
| CENG-04  | プレイヤー無言時のDIALスキップ | PlayerText=nil                                  | `BuildTranslationRequests(data, config)` | DIAL FULLのリクエストは生成されない                          |
| CENG-05  | トピック名抽出優先順位         | TopicText=有, NAM1=有                           | `BuildTranslationRequests(data, config)` | TopicNameが優先的にTopicTextとなる                           |
| CENG-06  | クエスト生成(CNAM, NNAM)       | Stages=3件, Objectives=2件のQuest               | `BuildTranslationRequests(data, config)` | QUST CNAM(x3), QUST NNAM(x2)のリクエストがそれぞれ生成される |
| CENG-07  | アイテムDESC生成               | TypeHint="Weapon", Desc="A sharp blade"         | `BuildTranslationRequests(data, config)` | `Weapon DESC` リクエストが要求通りに生成される               |
| CENG-08  | 書籍チャンク処理               | TypeHint="Book", 閾値を超過する長文のText       | `BuildTranslationRequests(data, config)` | 長文がBook DESC複数リクエストへと自動的に分割される          |
| CENG-09  | 翻訳済みスキップ               | 既に日本語が含まれるテキスト                    | `BuildTranslationRequests(data, config)` | 当該テキストに対するリクエスト生成から除外される             |
| CENG-10  | Lookupエラーの耐性             | TermLookup.Searchがエラーを返す意図的モック設定 | `BuildTranslationRequests(data, config)` | ログに該当エラーが出力されパニックせず後続の処理は継続する   |
| CENG-11  | 空データの入力                 | ExtractedDataが空のセット                       | `BuildTranslationRequests(data, config)` | 空配列 `[]` が返る。エラーなし                               |

---

## 3. 構造化ログとデバッグフロー (Log-based Debugging)

本スライスで不具合が発生した場合やテストが失敗した場合は、ステップ実行やユニットテストの追加による原因追及を行わず、実行生成物である構造化ログを用いたAIデバッグを徹底する。

### 3.1 テスト基盤でのログ準備
テストコードから Contract メソッドを呼び出す際は、サブテスト（Table-Drivenの各ケース）ごとに一意の TraceID を持つ `context.Context` を生成して引き回すこと。
各テスト実行時の `slog` の出力先はファイル（例: `logs/test_{timestamp}_ContextEngine.jsonl`）に記録するようルーティングする。

### 3.2 AIデバッグ専用プロンプトの定型化
障害発生時、またはテスト失敗時には、該当ジョブのログファイルをそのままAIに渡し、以下の定型プロンプトを用いたデバッグと修正指示を行う。

```text
以下はスライス「Context Engine Slice」の実行ログファイル（logs/{LogFilePath}）の内容である。
仕様書（openspec/specs/ContextEngineSlice/spec.md）の期待動作と比較し、乖離がある箇所を特定して修正コードを生成せよ。

--- 実行ログ ---
{ログファイル内容}

--- 期待される仕様 ---
{仕様書の該当セクション}
```
