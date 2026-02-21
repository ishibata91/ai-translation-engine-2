# 用語翻訳・Mod用語DB保存 テスト設計

## 1. ユニットテスト (Unit Tests)

### 1.1 `TermRequestBuilder` (リクエスト生成)
*   **対象**: `BuildRequests` メソッド
*   **テストケース**:
    *   正常系: NPC、Item、Magic、Message、Location、Questを含む `ExtractedData` を渡し、対象レコードタイプのみが `TermTranslationRequest` として生成されること。
    *   正常系 (フィルタリング): `TermRecordConfig` で指定されたレコードタイプのみが抽出され、対象外（例: `INFO:NAM1` 等の会話テキスト）は除外されること。
    *   正常系 (日本語スキップ): ソーステキストが既に日本語を含む場合、リクエストが生成されないこと。
    *   エッジケース: `ExtractedData` が空（NPC/Item等が0件）の場合、空のスライスが返ること。
    *   エッジケース: ソーステキストが空文字列のレコードはスキップされること。
    *   正常系 (NPCペアリング): 同一EditorIDの `NPC_:FULL` と `NPC_:SHRT` が1つの `TermTranslationRequest` にペアリングされ、`ShortName` フィールドにSHRTのテキストが設定されること。
    *   正常系 (NPC FULLのみ): SHRTが存在しないNPCは、`ShortName` が空の単独リクエストとして生成されること。
    *   エッジケース (SHRTのみNPC): FULLが存在せずSHRTのみのNPCは、SHRTを`SourceText`とする単独リクエストとして生成されること。

### 1.2 `TermRecordConfig` (対象レコード判定)
*   **対象**: `IsTarget` メソッド
*   **テストケース**:
    *   正常系: 許可リストに含まれるレコードタイプ（例: `NPC_:FULL`）に対して `true` を返すこと。
    *   正常系: 許可リストに含まれないレコードタイプ（例: `INFO:NAM1`）に対して `false` を返すこと。

### 1.3 `TermDictionarySearcher` (辞書検索)
*   **対象**: `SearchExact`, `SearchKeywords`, `SearchNPCPartial`, `SearchBatch` メソッド
*   **テスト環境**: In-Memory SQLite (`file::memory:?cache=shared`) を利用。テスト用の辞書データおよびNPC名FTS5データを事前にINSERTする。
*   **テストケース**:
    *   正常系 (完全一致): `"Auriel's Bow"` で検索し、辞書に登録済みの `"アーリエルの弓"` が返ること。
    *   正常系 (キーワード検索): `"Iron Sword of Fire"` からキーワード `["Iron Sword", "Fire"]` を抽出し、それぞれの既訳が返ること。
    *   正常系 (NPC部分一致 - 苗字): 辞書に `"Jon Battle-Born"` が登録されている状態で、キーワード `"Battle-Born"` で `SearchNPCPartial` を実行し、`"Jon Battle-Born"` のエントリが返ること。
    *   正常系 (NPC部分一致 - ファーストネーム): 辞書に `"Ulfric Stormcloak"` が登録されている状態で、キーワード `"Ulfric"` で部分一致検索し、フルネームのエントリが返ること。
    *   正常系 (NPC部分一致 - 消費済みキーワード除外): 完全一致で既に消費されたキーワードは `SearchNPCPartial` の対象から除外されること。
    *   正常系 (NPCレコード自体の翻訳時): `isNPC=true` の場合、全キーワードが部分一致検索の対象となること。
    *   正常系 (バッチ検索): 複数のソーステキストを一括検索し、各テキストに対応する参照用語マップが返ること。
    *   正常系 (該当なし): 辞書に存在しないテキストで検索した場合、空の結果が返ること。
    *   正常系 (ステミングフォールバック): 辞書に `"Sword"` が登録済みの状態で、キーワード `"Swords"` で検索した場合、完全一致はヒットせず、ステム化フォールバックで `"Sword"` がヒットすること。
    *   正常系 (所有格ステミング): 辞書に `"Auriel"` が登録済みの状態で、キーワード `"Auriel's"` で検索した場合、所有格除去+ステム化で `"Auriel"` がヒットすること。
    *   正常系 (ステム低優先): 完全一致結果とステム一致結果が競合する場合、完全一致が優先されること。
    *   正常系 (追加辞書): 追加辞書DBパスが指定されている場合、メイン辞書と追加辞書の両方から検索されること。
    *   エッジケース: 辞書DBが空の場合、エラーなく空の結果が返ること。
    *   エッジケース: `npc_terms_fts` テーブルが存在しない場合、`SearchNPCPartial` がエラーなく空の結果を返すこと（フォールバック）。

### 1.3.1 `KeywordStemmer` (ステミング)
*   **対象**: `Stem`, `StripPossessive`, `StemKeywords` メソッド
*   **テストケース**:
    *   正常系 (複数形): `"Swords"` → `"sword"` にステム化されること。
    *   正常系 (所有格): `"Auriel's"` → `StripPossessive` で `"Auriel"` になり、その後ステム化されること。
    *   正常系 (変化なし): `"Dragon"` のステムが `"dragon"` と同一の場合、`StemKeywords` の結果マップから除外されること。
    *   エッジケース: 空文字列や単一文字に対してパニックしないこと。

### 1.3.2 `GreedyLongestMatcher` (貪欲最長一致フィルタリング)
*   **対象**: `Filter` メソッド
*   **テストケース**:
    *   正常系 (最長一致優先): ソーステキスト `"Broken Tower Redoubt"` に対し、候補 `{"Broken Tower Redoubt": "ブロークン・タワー・リダウト", "Broken": "壊れた"}` がある場合、`"Broken Tower Redoubt"` のみが採用され `"Broken"` は抑制されること。
    *   正常系 (非重複複数マッチ): ソーステキスト `"Iron Sword and Iron Shield"` に対し、`"Iron Sword"` と `"Iron Shield"` の両方が採用されること。
    *   正常系 (区間重複排除): ソーステキスト `"Daedric Sword"` に対し、`"Daedric Sword"` と `"Sword"` が候補の場合、`"Daedric Sword"` のみ採用されること。
    *   エッジケース: 候補が空の場合、空の結果が返ること。
    *   エッジケース: ソーステキストに候補が一切含まれない場合、空の結果が返ること。

### 1.4 `ModTermStore` (Mod用語DB永続化)
*   **対象**: `InitSchema`, `SaveTerms`, `GetTerm`, `Clear` メソッド
*   **テスト環境**: In-Memory SQLite を利用。
*   **テストケース**:
    *   正常系 (スキーマ初期化): `InitSchema` 実行後、`mod_terms` テーブル、`mod_terms_fts` 仮想テーブル、`npc_terms_fts` 仮想テーブル、および同期トリガーが作成されていること。
    *   正常系 (保存): 複数の `TermTranslationResult` を渡し、エラーなくDBへ保存されること。`GetTerm` で保存した用語が取得できること。
    *   正常系 (UPSERT): 同一の `(original_en, record_type)` に対して再度 `SaveTerms` を実行した場合、既存レコードが更新されること。
    *   正常系 (FTS同期): `SaveTerms` 後、`mod_terms_fts` テーブルからFTS5検索で用語が取得できること。
    *   正常系 (NPC FTS): NPC名レコード（`record_type = "NPC_ FULL"`, `"NPC_ SHRT"`）が `npc_terms_fts` にも登録されること。
    *   正常系 (クリア): `Clear` 実行後、`mod_terms` テーブルおよびFTSテーブルが空になること。
    *   異常系: DBコネクション切断状態でのエラーハンドリング確認。
    *   エッジケース: `Status` が `"failed"` の結果は保存対象から除外されること。

### 1.5 `TermPromptBuilder` (プロンプト生成)
*   **対象**: `BuildSystemPrompt`, `BuildUserPrompt` メソッド
*   **テストケース**:
    *   正常系: レコードタイプ `"WEAP:FULL"` に対して武器名翻訳用のシステムプロンプトが生成されること。
    *   正常系: 参照用語リストが含まれるリクエストに対して、ユーザープロンプトに参照用語セクションが含まれること。
    *   エッジケース: 参照用語が空の場合、参照用語セクションが省略されること。

### 1.6 `TermTranslatorImpl` (強制翻訳)
*   **対象**: 強制翻訳（Forced Translation）ロジック
*   **テストケース**:
    *   正常系: 辞書に完全一致する既訳がある場合、LLMを呼び出さずに辞書訳が `TranslatedText` に設定され、`Status` が `"cached"` となること。
    *   正常系: 辞書に一致がない場合、LLMが呼び出されること。

## 2. 統合テスト (Integration Tests)

### 2.1 `TermTranslatorImpl` (エンドツーエンド)
*   **対象**: `TranslateTerms` メソッド
*   **テスト環境**: In-Memory SQLite + モックLLMClient。
*   **テストケース**:
    *   正常系: NPC名・アイテム名を含む `ExtractedData` を渡し、全用語が翻訳され、Mod用語DBに保存されること。
    *   正常系 (NPCペア翻訳): FULLとSHRTを持つNPCが同一LLMリクエストで翻訳され、結果がFULL・SHRTそれぞれ個別のレコードとしてMod用語DBに保存されること。SHRTの訳がFULLの訳と整合していること。
    *   正常系 (混合): 辞書に既訳がある用語（Forced Translation）とない用語（LLM翻訳）が混在する場合、それぞれ正しく処理されること。
    *   正常系 (貪欲部分一致の参照用語): NPC苗字の部分一致で得られた参照用語が、LLMプロンプトの `reference_terms` に含まれ、かつ強制翻訳の対象にはならないこと。
    *   正常系 (並列): 複数のリクエストが並列で処理され、全結果が正しく返却されること。
    *   正常系 (進捗通知): `ProgressNotifier` のモックを注入し、翻訳の進捗が正しく通知されること。
    *   異常系 (LLM部分失敗): 一部のLLM呼び出しが失敗しても、プロセス全体は停止せず、失敗レコードのみ `Status = "failed"` となること。成功した用語はMod用語DBに保存されること。
    *   異常系 (全LLM失敗): 全てのLLM呼び出しが失敗した場合でも、Forced Translationで解決された用語はMod用語DBに保存されること。

### 2.2 HTTP ハンドラー (`ProcessManager`)
*   **対象**: `HandleTermTranslation` メソッド
*   **テスト環境**: `net/http/httptest` を利用。
*   **テストケース**:
    *   正常系: POST リクエストでPass 1実行を開始し、ステータス `200 OK` および翻訳結果サマリーが返却されること。
    *   異常系: `ExtractedData` が未ロード状態でリクエストした場合、ステータス `400 Bad Request` を返すこと。

## 3. UI動作テスト (Manual / E2E)

*   **フロントエンド**: React UI（構築予定）より、ロード済みのModデータに対してPass 1（用語翻訳）を実行。
*   **検証項目**:
    *   用語翻訳の進捗バーがリアルタイムで更新されるか。
    *   翻訳完了後、Mod用語DBファイルが生成され、用語が正しく保存されているか。
    *   大量の用語（例: 1000件以上のNPC名・アイテム名）でもタイムアウトせずに処理が完了するか。
    *   Forced Translationにより、公式DLC辞書に存在する用語はLLM呼び出しなしで即座に解決されるか。
    *   Pass 1完了後、Pass 2（本文翻訳）でMod用語DBが `additional_db_paths` として正しく参照されるか。
