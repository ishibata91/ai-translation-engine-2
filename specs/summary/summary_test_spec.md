# 要約ジェネレータ テスト設計 (SummaryGeneratorSlice Test Spec)

本設計は `architecture.md` セクション 6（テスト戦略）および セクション 7（構造化ログ基盤）に厳密に準拠し、個別関数の細粒度なユニットテストを作成せず、Vertical Slice の Contract に対する網羅的なパラメタライズドテスト（Table-Driven Test）を定義する。

## 1. テスト方針

1. **細粒度ユニットテストの排除**: 内部のモデルの機能分割（`CacheKeyHasher`, `SQLiteSummaryStore`, `SummaryGeneratorImpl`内の個別関数 など）ごとの細粒度なユニットテストは作成しない。
2. **網羅的パラメタライズドテスト**: Contract（`GenerateDialogueSummaries`, `GenerateQuestSummaries`）を入力境界とし、入力テキストの受け取りからキャッシュ判定、LLMによる要約生成、そしてインメモリDB（SQLite `:memory:` 設定のStore）への保存ないしキャッシュヒットまでのフロー全体を Table-Driven Test として検証する。API通信は `net/http/httptest` モックサーバーまたはモックインフラを利用する。
3. **構造化ログの強制検証**: テスト実行時においても、必ず `context.Context` （独自の TraceID を内包）を引き回す。

---

## 2. パラメタライズドテストケース (Table-Driven Tests)

各Contractに対する入力（初期状態 + アクション）と期待されるアウトプット（戻り値 + 変更後の状態）を表として定義し、ループ内で検証する。

### 2.1 Summary.GenerateDialogueSummaries / GenerateQuestSummaries 統合テスト
`DialogueGroupInput` または `QuestInput` の一覧を受け取り、必要に応じてLLMを呼び出し要約を生成してキャッシュへ保存、あるいはキャッシュから結果を引く動作を一貫してテストする。

| ケースID | 目的                                                   | 初期状態 (入力/設定/モック等)                                                                                                | アクション (関数呼出)                                                                                 | 期待される結果 (出力 / 状態)                                                                                                                         |
| :------- | :----------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------- |
| SGN-01   | 正常系: キャッシュミスと生成(会話)                     | 空のインメモリDB。<br>モックLLMが要約テキストを返却する状態。<br>入力: 1件の `DialogueGroupInput` (複数行のテキストを含む)。 | `GenerateDialogueSummaries(...)`                                                                      | エラーなし。<br>LLMが呼ばれ、要約結果が返り(`CacheHit=false`)、結果がSQLiteに保存されること。                                                        |
| SGN-02   | 正常系: キャッシュヒット(会話)                         | インメモリDBにケースID: SGN-01と同じ `DialogueGroupInput` に対応する要約レコードが存在する状態。                             | `GenerateDialogueSummaries(...)`                                                                      | エラーなし。<br>LLMは呼ばれず、DBに保存されたテキストが返却される(`CacheHit=true`)こと。                                                             |
| SGN-03   | 正常系: 内容変更によるキャッシュ無効化と再生成         | インメモリDBに要約レコードが存在するが、入力の `DialogueGroupInput` のテキスト行が一部異なる状態。                           | `GenerateDialogueSummaries(...)`                                                                      | エラーなし。<br>LLMが呼ばれ(`CacheHit=false`)、新しい要約でDBが上書き(UPSERT)されること。                                                            |
| SGN-04   | 正常系: クエスト要約の生成(複数ステージ)               | 空のインメモリDB。<br>モックLLMが応答する状態。<br>入力: 順不同のステージテキストを含む `QuestInput` 。                      | `GenerateQuestSummaries(...)`                                                                         | エラーなし。<br>内部で時系列順にプロンプトが構築されてLLMが呼ばれ、結果がSQLiteに保存されること。                                                    |
| SGN-05   | 正常系: 空入力のスキップと空文字返却                   | 空のインメモリDB。<br>入力: `Lines` が空の `DialogueGroupInput`、または `StageTexts` が空の `QuestInput`。                   | `GenerateDialogueSummaries(...)` / `GenerateQuestSummaries(...)`                                      | エラーなし。<br>LLMは呼ばれず、空文字が返り、DBへの保存も行われないこと。                                                                            |
| SGN-06   | 異常系: LLM失敗時の継続処理                            | モックLLMがエラーを返す状態。<br>入力: 複数の要約対象入力群。                                                                | 該当メソッド呼出                                                                                      | 処理全体はパニックせずに完了すること。<br>該当する結果は空文字となり(エラーログが記録される前提)、他の正常な入力群の生成結果は正しく返却されること。 |
| SGN-07   | 正常系: 異なるソースファイルの独立したキャッシュDB作成 | 初期状態なし。                                                                                                               | ソースA, ソースB それぞれのDB生成メソッドを経由して別々のファイルを指定し、同じレコードIDで保存・取得 | 互いに干渉せず、ソースAの要約保存がソースBのDBから検索(ヒット)できないことが確認できること。                                                         |

---

## 3. 構造化ログとデバッグフロー (Log-based Debugging)

本スライスで不具合が発生した場合やテストが失敗した場合は、ステップ実行やユニットテストの追加による原因追及を行わず、実行生成物である構造化ログを用いたAIデバッグを徹底する。

### 3.1 テスト基盤でのログ準備
テストコードから Contract メソッドを呼び出す際は、サブテスト（Table-Drivenの各ケース）ごとに一意の TraceID を持つ `context.Context` を生成して引き回すこと。
各テスト実行時の `slog` の出力先はファイル（例: `logs/test_{timestamp}_SummaryGeneratorSlice.jsonl`）に記録するようルーティングする。

### 3.2 AIデバッグプロンプトテンプレート
テスト失敗時やバグ発生時は、出力されたログと本仕様書を用いてAIにデバッグを依頼する。

```text
以下はスライス「SummaryGeneratorSlice」の実行ログファイル（logs/test_XXXXX_SummaryGeneratorSlice.jsonl）の内容である。
仕様書（openspec/specs/SummaryGeneratorSlice/spec.md および 本テスト設計）の期待動作と比較し、乖離がある箇所を特定して修正コードを生成せよ。

--- 実行ログ ---
{ログファイル内容}

--- 期待される仕様 ---
{仕様書の該当セクション}
```
