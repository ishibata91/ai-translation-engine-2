# NPCペルソナ生成 テスト設計

## 1. ユニットテスト (Unit Tests)

### 1.1 `DialogueCollector` (会話データ収集)
*   **対象**: `CollectByNPC` メソッド
*   **テストケース**:
    *   正常系: 複数のNPC（SpeakerID）を含む `ExtractedData` を渡し、SpeakerIDごとにグルーピングされた `NPCDialogueData` が返ること。
    *   正常系 (上限100件): 1つのNPCに150件の会話がある場合、重要度スコアリングにより上位100件のみ収集されること。
    *   正常系 (100件以下): 1つのNPCの会話が100件以下の場合、スコアリングなしで全件がそのまま収集されること。
    *   正常系 (スコア順選別): 重要度スコアの高い会話が優先的に選別され、低スコアの会話が除外されること。
    *   正常系 (同スコア時Order昇順): 同一重要度スコアの会話は `Order` の昇順で選択されること。
    *   正常系 (日本語テキスト選択): 選別後の出力が日本語テキストであること。日本語テキストが存在しない場合は英語原文がフォールバックとして使用されること。
    *   正常系 (NPC属性解決): `SpeakerID` に対応するNPCモデルから `Race`, `Sex`, `VoiceType` が正しく設定されること。
    *   正常系 (NPC属性なし): NPC属性が取得できない場合でも、会話データのみで `NPCDialogueData` が生成されること（属性フィールドは空文字）。
    *   エッジケース (SpeakerID nil): `SpeakerID` がnilの `DialogueResponse` は収集対象外となること。
    *   エッジケース (SpeakerID 空文字): `SpeakerID` が空文字の `DialogueResponse` は収集対象外となること。
    *   エッジケース (空データ): `ExtractedData` に `DialogueGroup` が0件の場合、空のスライスが返ること。

### 1.2 `ImportanceScorer` (重要度スコアリング)
*   **対象**: `Score` メソッド (`ImportanceScorerImpl`)
*   **テストケース**:
    *   正常系 (固有名詞スコア): 英語テキスト `"Go to Whiterun and find Jon Battle-Born"` に対して、Mod用語DBに `"Whiterun"`, `"Jon Battle-Born"` が登録済みの場合、`proper_noun_hits = 2` が算出されること。
    *   正常系 (Word Boundaryマッチ): `"Winterhold"` が登録済みの状態で、テキスト `"Winter is coming"` に対して `"Winter"` が部分一致しないこと（`\b` による境界マッチ）。
    *   正常系 (感情スコア - 句読点): テキスト `"Stop! What are you doing?!"` に対して、`!` x2 + `?` x1 = 3 が `emotion_indicators` に含まれること。
    *   正常系 (感情スコア - ALL CAPS): テキスト `"You NEVER listen to me"` に対して、全大文字単語 `"NEVER"` が1件検出されること。
    *   正常系 (感情スコア - 強い単語): テキスト `"I will kill you, you fool"` に対して、強い単語リスト (`"kill"`, `"fool"`) から2件マッチすること。
    *   正常系 (base_priority - クエスト関連): `QuestID` が非nilの場合、`base_priority = +10` が加算されること。
    *   正常系 (base_priority - サービス分岐): `IsServicesBranch == true` の場合、`base_priority = +0` となること。
    *   正常系 (base_priority - その他): `QuestID` がnilかつ `IsServicesBranch == false` の場合、`base_priority = +5` となること。
    *   正常系 (スコア集計): `proper_noun_hits=2, W_noun=2, emotion_indicators=3, W_emotion=1, base_priority=10` の場合、`importance_score = 2*2 + 3*1 + 10 = 17` となること。
    *   エッジケース: 空文字列に対してスコア `0 + base_priority` が返ること。
    *   エッジケース: 固有名詞・感情指標が一切ない平文に対して `base_priority` のみが返ること。

### 1.3 `TokenEstimator` (トークン推定)
*   **対象**: `Estimate` メソッド (`SimpleTokenEstimator`)
*   **テストケース**:
    *   正常系: 英語テキスト `"Hello world"` (11文字) に対して `11/4 = 2` (切り捨て) が返ること。
    *   正常系: 長文テキスト (400文字) に対して `100` が返ること。
    *   エッジケース: 空文字列に対して `0` が返ること。
    *   エッジケース: 1文字のテキストに対して `0` (切り捨て) が返ること。

### 1.4 `ContextEvaluator` (コンテキスト長評価)
*   **対象**: `Evaluate` メソッド
*   **テストケース**:
    *   正常系 (上限内): 合計トークン数がコンテキストウィンドウ上限以下の場合、`ExceedsLimit = false` で全会話データがそのまま返ること。
    *   正常系 (超過 - 削減): 合計トークン数が上限を超過する場合、会話データが削減され上限内に収まること。削減後の会話件数が10件以上であること。
    *   正常系 (超過 - 最低10件保証): 大量のトークンを持つ会話データで、削減しても10件未満にならないこと。10件で上限を超過する場合は警告フラグが設定されること。
    *   正常系 (トークン計算): `InputTokens` にシステムプロンプトオーバーヘッドが加算されていること。`TotalTokens` が `InputTokens + MaxOutputTokens` であること。
    *   エッジケース: 会話データが0件の場合、`InputTokens` がシステムプロンプトオーバーヘッドのみとなること。

### 1.5 `PersonaStore` (ペルソナDB永続化)
*   **対象**: `InitSchema`, `SavePersona`, `GetPersona`, `Clear` メソッド
*   **テスト環境**: In-Memory SQLite (`file::memory:?cache=shared`) を利用。
*   **テストケース**:
    *   正常系 (スキーマ初期化): `InitSchema` 実行後、`npc_personas` テーブルが作成されていること。
    *   正常系 (保存): `PersonaResult` を渡し、エラーなくDBへ保存されること。`GetPersona` で保存したペルソナが取得できること。
    *   正常系 (UPSERT): 同一 `speaker_id` に対して再度 `SavePersona` を実行した場合、既存レコードが更新されること（`persona_text`, `dialogue_count`, `estimated_tokens`, `updated_at` が更新）。
    *   正常系 (クリア): `Clear` 実行後、`npc_personas` テーブルが空になること。
    *   正常系 (該当なし): 存在しない `speaker_id` で `GetPersona` を実行した場合、空文字列が返ること。
    *   異常系: DBコネクション切断状態でのエラーハンドリング確認。
    *   エッジケース: `Status` が `"failed"` の結果は保存対象から除外されること。

### 1.6 `PersonaPromptBuilder` (プロンプト生成)
*   **対象**: `BuildSystemPrompt`, `BuildUserPrompt` メソッド
*   **テストケース**:
    *   正常系: システムプロンプトにペルソナ生成の目的（Skyrim Modの翻訳における口調決定）と出力フォーマット指定（性格特性・話し方の癖・背景設定の3セクション）が含まれること。
    *   正常系 (NPC属性あり): NPC属性（種族・性別・声の種類）を持つ `NPCDialogueData` に対して、ユーザープロンプトにNPC属性セクションが含まれること。
    *   正常系 (NPC属性なし): NPC属性が空の `NPCDialogueData` に対して、ユーザープロンプトにNPC属性セクションが省略され、会話データのみが含まれること。
    *   正常系 (会話テキスト一覧): ユーザープロンプトに収集した会話テキスト一覧（最大100件）が含まれること。
    *   エッジケース: 会話データが1件のみの場合でも、正しいプロンプトが生成されること。

## 2. 統合テスト (Integration Tests)

### 2.1 `NPCPersonaGeneratorImpl` (エンドツーエンド)
*   **対象**: `GeneratePersonas` メソッド
*   **テスト環境**: In-Memory SQLite + モックLLMClient。
*   **テストケース**:
    *   正常系: 複数のNPCを含む `ExtractedData` を渡し、全NPCのペルソナが生成され、`npc_personas` テーブルに保存されること。
    *   正常系 (重要度スコアリング): 150件の会話を持つNPCに対して、重要度スコアリングにより上位100件が選別され、選別後のデータでペルソナが正常に生成されること。
    *   正常系 (コンテキスト長評価): コンテキストウィンドウ上限を超過するNPCの会話データが自動的に削減され、ペルソナが正常に生成されること。
    *   正常系 (会話0件スキップ): 会話データが0件のNPCはペルソナ生成がスキップされ、結果に含まれないこと。
    *   正常系 (並列): 複数のNPCのペルソナが並列で生成され、全結果が正しく返却されること。
    *   正常系 (進捗通知): `ProgressNotifier` のモックを注入し、ペルソナ生成の進捗（完了NPC数/対象NPC総数）が正しく通知されること。
    *   正常系 (UPSERT): 同一NPCに対して再度 `GeneratePersonas` を実行した場合、既存ペルソナが上書きされること。
    *   異常系 (LLM部分失敗): 一部のLLM呼び出しが失敗しても、プロセス全体は停止せず、失敗NPCのみ `Status = "failed"` となること。成功したNPCのペルソナは `npc_personas` テーブルに保存されること。
    *   異常系 (全LLM失敗): 全てのLLM呼び出しが失敗した場合、空の結果が返り、エラーが適切に記録されること。

### 2.2 Pass 2 ペルソナ参照
*   **対象**: `PersonaStore.GetPersona` メソッド（Pass 2からの参照）
*   **テスト環境**: In-Memory SQLite に事前にペルソナデータをINSERT。
*   **テストケース**:
    *   正常系: `SpeakerID` で完全一致検索し、保存済みの `persona_text` が返ること。
    *   正常系 (フォールバック): ペルソナが存在しない `SpeakerID` で検索した場合、空文字列が返り、呼び出し元が従来の種族・声タイプベース推定にフォールバックできること。

### 2.3 HTTP ハンドラー (`ProcessManager`)
*   **対象**: `HandlePersonaGeneration` メソッド
*   **テスト環境**: `net/http/httptest` を利用。
*   **テストケース**:
    *   正常系: POST リクエストでペルソナ生成を開始し、ステータス `200 OK` および生成結果サマリーが返却されること。
    *   異常系: `ExtractedData` が未ロード状態でリクエストした場合、ステータス `400 Bad Request` を返すこと。

## 3. UI動作テスト (Manual / E2E)

*   **フロントエンド**: React UI（構築予定）より、ロード済みのModデータに対してPass 1（ペルソナ生成）を実行。
*   **検証項目**:
    *   ペルソナ生成の進捗バー（完了NPC数/対象NPC総数）がリアルタイムで更新されるか。
    *   生成完了後、`npc_personas` テーブルにペルソナが正しく保存されているか。
    *   コンテキスト長超過が発生したNPCについて、警告がUIに表示されるか。
    *   大量のNPC（例: 200体以上）でもタイムアウトせずに処理が完了するか。
    *   Pass 2（本文翻訳）で、生成されたペルソナがシステムプロンプトの話者プロファイルに正しく挿入されるか。
    *   ペルソナが存在しないNPCについて、従来の種族・声タイプベース推定にフォールバックされるか。
