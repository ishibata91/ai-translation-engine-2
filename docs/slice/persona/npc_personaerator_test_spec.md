# NPC ペルソナ生成 テスト仕様

本設計は `standard_test_spec.md` と `log-guide.md` に厳密に準拠し、個別関数の細粒度なユニットテストを作成せず、Vertical Slice の Contract に対する網羅的なパラメタライズドテスト（Table-Driven Test）を定義する。

## 1. テスト方針

1. **細粒度ユニットテストの排除**: 内部のモデルの機能分割（`DialogueCollector`, `ImportanceScorer`, `TokenEstimator`, `ContextEvaluator`, `PersonaPromptBuilder` など）ごとの細粒度なユニットテストは作成しない。
2. **網羅的パラメタライズドテスト**: Contract（`GeneratePersonas`）を入力境界とし、`ExtractedData` を受け取ってから、内部での収集・スコアリング・API呼び出しを経て、インメモリDB（`PersonaStore`相当）へ結果が永続化されるまでの一連のフローを Table-Driven Test として検証する。API通信は `net/http/httptest` モックサーバーまたはモックインフラを利用する。
3. **構造化ログの強制検証**: テスト実行時においても、必ず `context.Context` （独自の TraceID を内包）を引き回す。

---

## 2. パラメタライズドテストケース (Table-Driven Tests)

各Contractに対する入力（初期状態 + アクション）と期待されるアウトプット（戻り値 + 変更後の状態）を表として定義し、ループ内で検証する。

### 2.1 PersonaGenerator.GeneratePersonas 統合テスト
`ExtractedData` に含まれる `DialogueGroup` / `DialogueResponse` からNPCを識別し、会話収集、LLMペルソナ生成、およびSQLiteへの保存までを一貫してテストする。

| ケースID | 目的                                                    | 初期状態 (入力/設定/モック等)                                                                                                                       | アクション (関数呼出)         | 期待される結果 (出力 / 状態)                                                                                                                                        |
| :------- | :------------------------------------------------------ | :-------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| PGN-01   | 正常系: 複数NPCの基本生成と保存                         | 空のインメモリDB。<br>モックLLMがペルソナテキストを返却する状態。<br>入力: 2人のNPCの会話データ(各数件)が含まれる `ExtractedData`。                 | `GeneratePersonas(ctx, data)` | エラーなし。<br>2人分のペルソナが生成され、インメモリDBの `npc_personas` テーブルに正しく保存されること。                                                           |
| PGN-02   | 正常系: 会話上限(100件)超えの重要度スコアリング選別検証 | モックLLMが正常応答する状態。<br>入力: 1人のNPCで150件の会話を持つデータ（うち一部は特定の感情単語や固有名詞を意図的に含むよう設定）。              | `GeneratePersonas(ctx, data)` | エラーなし。<br>スコアリングによる選別を経て上位100件と推定される内容がモックに送られ、生成結果がDBに保存されること。                                               |
| PGN-03   | 正常系: トークン制限超過時の圧縮と警告                  | モックLLMが正常応答する状態。<br>入力: 1人のNPCについて、合計トークン数が設定されたコンテキスト上限(例:1k)を遥かに超える超長文の会話データ。        | `GeneratePersonas(ctx, data)` | エラーなし。<br>内部で会話データが削減されてLLMに送られ、生成完了後にログ（またはコールバック等）でコンテキスト超過の警告が発せられていること。DBに保存されること。 |
| PGN-04   | 正常系: 会話0件のNPCのスキップ                          | 空のインメモリDB。<br>モックLLMが正常応答する状態。<br>入力: SpeakerIDはあるがテキストが空、あるいは該当するDialogueが存在しない(0件)状態のデータ。 | `GeneratePersonas(ctx, data)` | エラーなし。<br>LLMは呼ばれず、DBにも該当NPCのペルソナは保存されない（処理がスキップされる）こと。                                                                  |
| PGN-05   | 正常系: 同一NPCの差分更新(UPSERT)                       | 既存ペルソナが存在するインメモリDB。<br>入力: 既存と同じ SpeakerID を持つ新たな会話データ。                                                         | `GeneratePersonas(ctx, data)` | エラーなし。<br>DBの既存レコードが新しい生成結果として更新(UPSERT)されていること。                                                                                  |
| PGN-06   | 異常系: LLM部分失敗でもプロセス継続                     | モックLLMがNPC[A]には正常、NPC[B]にはエラーを返す状態。<br>入力: NPC[A], NPC[B] の会話データ。                                                      | `GeneratePersonas(ctx, data)` | 処理全体はパニックせず完了すること。<br>NPC[A]のペルソナのみがDBに保存され、NPC[B]は保存されない（また失敗リスト等に含まれる）こと。                                |

### 2.2 PersonaStore 参照統合テスト (Pass 2境界用)
ペルソナ生成後、Pass 2（本文翻訳）Slice がペルソナを参照する際の Read API の振る舞いを検証する。

| ケースID | 目的                             | 初期状態 (入力/設定/モック等)                       | アクション (関数呼出)            | 期待される結果 (出力 / 状態)                                                                                         |
| :------- | :------------------------------- | :-------------------------------------------------- | :------------------------------- | :------------------------------------------------------------------------------------------------------------------- |
| PSR-01   | 正常系: ペルソナ参照             | DBに SpeakerID="NPC_01" のペルソナが存在する状態。  | `GetPersona(ctx, "NPC_01")`      | エラーなし。<br>保存済みのペルソナテキストが取得できること。                                                         |
| PSR-02   | 正常系: 該当なしのフォールバック | DBが空、または該当するSpeakerIDのデータがない状態。 | `GetPersona(ctx, "NPC_UNKNOWN")` | エラーなし。<br>空文字列（または該当なしを示す仕様準拠の戻り値）が返り、呼び出し元がフォールバック処理を行えること。 |

### 2.3 Translation Flow persona phase 統合テスト
translation flow の persona phase が、既存 Master Persona の再利用、候補統合、retry を正しく扱うことを検証する。

| ケースID | 目的                                                | 初期状態 (入力/設定/モック等)                                                                                         | アクション (関数呼出)                                     | 期待される結果 (出力 / 状態)                                                                                                                                 |
| :------- | :-------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------- | :-------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| TFP-01   | preview と execute が既存 final persona を同じく除外する | translation input に NPC[A], NPC[B] が含まれ、`master_persona_artifact` には NPC[A] の final persona が存在する状態。 | `ListPersonaTargets` -> `RunTranslationFlowPersonaPhase` | NPC[A] は preview で `既存 Master Persona` として表示され、request に含まれないこと。NPC[B] だけが生成対象となること。                                   |
| TFP-02   | 全件再利用では no-op 完了する                       | translation input の全候補が `master_persona_artifact` final に存在する状態。                                        | `RunTranslationFlowPersonaPhase`                          | エラーなし。request 0 件。runtime を呼ばずに phase が完了し、summary が `新規生成 0 件` と `再利用済み` を返すこと。                                      |
| TFP-03   | 同一 lookup key の重複候補を 1 件へ統合する         | 同じ `source_plugin + speaker_id` を持つ NPC が複数 source file に現れる translation input。                         | `ListPersonaTargets`                                      | preview 上の候補は 1 件だけになり、重複 request 計画が作られないこと。                                                                                       |
| TFP-04   | retry は failed または未生成候補だけを再送する      | 初回実行で NPC[A] は成功、NPC[B] は失敗した phase 状態。                                                              | 再度 `RunTranslationFlowPersonaPhase`                     | NPC[A] は再送されず、NPC[B] だけが再 request 化されること。成功済み final persona が保持されること。                                                       |
| TFP-05   | lookup key 正規化が Master Persona と一致する       | `source_plugin` 欠損 input と source file ヒントを持つ候補、または `UNKNOWN` fallback が必要な候補を含む状態。        | `ListPersonaTargets` -> `RunTranslationFlowPersonaPhase` | preview と execute が同じ補完値で lookup し、既存 final persona の誤除外または誤再生成を起こさないこと。                                                   |

---

## 3. 構造化ログとデバッグフロー (Log-based Debugging)

本スライスで不具合が発生した場合やテストが失敗した場合は、ステップ実行やユニットテストの追加による原因追及を行わず、実行生成物である構造化ログを用いたAIデバッグを徹底する。

### 3.1 テスト基盤でのログ準備
テストコードから Contract メソッドを呼び出す際は、サブテスト（Table-Drivenの各ケース）ごとに一意の TraceID を持つ `context.Context` を生成して引き回すこと。
各テスト実行時の `slog` の出力先はファイル（例: `logs/test_{timestamp}_PersonaGenSlice.jsonl`）に記録するようルーティングする。

### 3.2 AIデバッグプロンプトテンプレート
テスト失敗時やバグ発生時は、出力されたログと本仕様書を用いてAIにデバッグを依頼する。

```text
以下はスライス「PersonaGenSlice」の実行ログファイル（logs/test_XXXXX_PersonaGenSlice.jsonl）の内容である。
仕様書（docs/slice/persona/spec.md および 本テスト設計）の期待動作と比較し、乖離がある箇所を特定して修正コードを生成せよ。

--- 実行ログ ---
{ログファイル内容}

--- 期待される仕様 ---
{仕様書の該当セクション}
```
