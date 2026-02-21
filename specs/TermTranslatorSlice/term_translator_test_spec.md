# 用語翻訳・Mod用語DB保存 テスト設計 (TermTranslatorSlice Test Spec)

本設計は `refactoring_strategy.md` セクション 6（テスト戦略）および セクション 7（構造化ログ基盤）に厳密に準拠し、個別関数の細粒度なユニットテストを作成せず、Vertical Slice の Contract に対する網羅的なパラメタライズドテスト（Table-Driven Test）を定義する。

## 1. テスト方針

1. **細粒度ユニットテストの排除**: 内部のモデルの機能分割（`TermRequestBuilder`, `TermRecordConfig`, `TermDictionarySearcher`, `KeywordStemmer`, `GreedyLongestMatcher`, `TermPromptBuilder` 等）ごとの細粒度なユニットテストは作成しない。
2. **網羅的パラメタライズドテスト**: Contract（`TranslateTerms`）を入力境界とし、`ExtractedData` を受け取ってから、対象レコードの抽出、辞書照合（完全一致/部分一致/ステミング）、LLMによる翻訳生成、そしてインメモリDB（`ModTermStore`相当）への保存までの一連のフローを Table-Driven Test として検証する。API通信はモックインフラを利用する。
3. **構造化ログの強制検証**: テスト実行時においても、必ず `context.Context` （独自の TraceID を内包）を引き回す。

---

## 2. パラメタライズドテストケース (Table-Driven Tests)

各Contractに対する入力（初期状態 + アクション）と期待されるアウトプット（戻り値 + 変更後の状態）を表として定義し、ループ内で検証する。

### 2.1 TermTranslator.TranslateTerms 統合テスト
`ExtractedData` に含まれる様々なレコード（NPC、アイテム等）から訳語を翻訳・生成し、Mod専用テーブルに保存される一連の処理をテストする。

| ケースID | 目的                                                 | 初期状態 (入力/設定/モック等)                                                                                                                                          | アクション (関数呼出)       | 期待される結果 (出力 / 状態)                                                                                                                                   |
| :------- | :--------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :-------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| TTN-01   | 正常系: 新規設定のLLM翻訳と保存                      | 空のMod用語DB、空の辞書DB（DictionaryBuilder由来）。<br>モックLLMが翻訳テキストを返却する状態。<br>入力: NPC(`NPC_:FULL`, `NPC_:SHRT`)と未訳Itemを含む `ExtractedData` | `TranslateTerms(ctx, data)` | エラーなし。<br>LLMが呼ばれ、NPCのFULLとSHRTがペア翻訳された結果と、Itemの結果が個別のレコードとしてMod用語DBに保存されること。                                |
| TTN-02   | 正常系: 強制翻訳 (完全一致)                          | 辞書DBに `Auriel's Bow` -> `アーリエルの弓` の完全一致エントリが存在。<br>入力: レコード `Auriel's Bow`。                                                              | `TranslateTerms(ctx, data)` | エラーなし。<br>LLMは呼ばれず、直接辞書の訳語が採用(`Status="cached"`)され、Mod用語DBに保存されること。                                                        |
| TTN-03   | 正常系: 貪欲部分一致とステミングによる参照用語       | 辞書DBに `Sword` -> `剣` などのエントリが存在。<br>入力: 辞書に完全一致はないが、`Swords`（複数形）を含むレコード。                                                    | `TranslateTerms(ctx, data)` | エラーなし。<br>ステミングと部分一致検索が機能し、LLMへのプロンプトに `Sword` の訳語が `reference_terms` として提供され、翻訳結果がDBに保存されること。        |
| TTN-04   | 正常系: NPC名部分一致検索(苗字/名前)                 | 辞書DBに `Jon Battle-Born` 等のエントリが存在し、`npc_terms_fts` が有効。<br>入力: `Battle-Born` のみを含むレコード。                                                  | `TranslateTerms(ctx, data)` | エラーなし。<br>FTSでNPC名が部分一致検索され、参照用語としてLLMに提供されること。                                                                              |
| TTN-05   | 正常系: 設定による特定レコードタイプのフィルタリング | `TermRecordConfig` にて特定のレコードタイプ(例:`INFO:NAM1`や翻訳済みのテキスト)が除外される設定。<br>入力: 除外対象レコードを含むデータ。                              | `TranslateTerms(ctx, data)` | エラーなし。<br>除外対象レコードのリクエストは生成されず、LLM呼び出しも行われないこと。対象レコードのみ翻訳・保存されること。                                  |
| TTN-06   | 正常系: 同一レコードのUPSERT更新                     | Mod用語DBに既存の翻訳レコードが存在する状態。<br>入力: 既存と同じ `(original_en, record_type)` を持つデータ。                                                          | `TranslateTerms(ctx, data)` | エラーなし。<br>DB内の既存レコードが新しい翻訳結果として更新（UPSERT）されること。                                                                             |
| TTN-07   | 異常系: LLM部分失敗時のプロセス継続                  | モックLLMが一部の用語リクエストに対してエラーを返す状態。                                                                                                              | `TranslateTerms(ctx, data)` | 処理全体はパニックせず完了すること。<br>成功した用語の翻訳およびForced Translationで解決した用語はDBに保存され、失敗分は除外(または失敗ステータス)されること。 |

---

## 3. 構造化ログとデバッグフロー (Log-based Debugging)

本スライスで不具合が発生した場合やテストが失敗した場合は、ステップ実行やユニットテストの追加による原因追及を行わず、実行生成物である構造化ログを用いたAIデバッグを徹底する。

### 3.1 テスト基盤でのログ準備
テストコードから Contract メソッドを呼び出す際は、サブテスト（Table-Drivenの各ケース）ごとに一意の TraceID を持つ `context.Context` を生成して引き回すこと。
各テスト実行時の `slog` の出力先はファイル（例: `logs/test_{timestamp}_TermTranslatorSlice.jsonl`）に記録するようルーティングする。

### 3.2 AIデバッグプロンプトテンプレート
テスト失敗時やバグ発生時は、出力されたログと本仕様書を用いてAIにデバッグを依頼する。

```text
以下はスライス「TermTranslatorSlice」の実行ログファイル（logs/test_XXXXX_TermTranslatorSlice.jsonl）の内容である。
仕様書（openspec/specs/TermTranslatorSlice/spec.md および 本テスト設計）の期待動作と比較し、乖離がある箇所を特定して修正コードを生成せよ。

--- 実行ログ ---
{ログファイル内容}

--- 期待される仕様 ---
{仕様書の該当セクション}
```
