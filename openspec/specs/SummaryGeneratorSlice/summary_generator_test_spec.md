# 要約ジェネレータ テスト設計

## テスト方針
- 全テストはモック（LLMClient, SummaryStore）を使用し、外部依存なしで実行可能とする。
- SQLiteSummaryStore のテストにはインメモリSQLite (`:memory:`) を使用する。

---

## 1. CacheKeyHasher テスト

### 1.1 同一入力で同一キーが生成される
- **入力**: recordID=`"DIAL001"`, lines=`["Hello", "World"]`
- **期待**: 2回呼び出しで同一の `cacheKey` が返る。

### 1.2 入力テキストが異なればキーが変わる
- **入力A**: recordID=`"DIAL001"`, lines=`["Hello"]`
- **入力B**: recordID=`"DIAL001"`, lines=`["Hello", "World"]`
- **期待**: 異なる `cacheKey` が返る。

### 1.3 レコードIDが異なればキーが変わる
- **入力A**: recordID=`"DIAL001"`, lines=`["Hello"]`
- **入力B**: recordID=`"DIAL002"`, lines=`["Hello"]`
- **期待**: 異なる `cacheKey` が返る。

### 1.4 キャッシュキー形式の検証
- **入力**: recordID=`"DIAL001"`, lines=`["test"]`
- **期待**: `"DIAL001|<sha256hex>"` 形式の文字列が返る。

---

## 2. SQLiteSummaryStore テスト

### 2.1 NewSQLiteSummaryStore: ソースファイル名からDBファイルが作成される
- **操作**: `NewSQLiteSummaryStore(tempDir, "Skyrim.esm")` を呼び出す。
- **期待**: `tempDir/Skyrim.esm_summary_cache.db` ファイルが作成される。

### 2.2 InitTable: テーブルとインデックスが作成される
- **操作**: `InitTable(ctx)` を呼び出す。
- **期待**: `summaries` テーブル、`idx_summaries_cache_key`、`idx_summaries_record_type` インデックスが存在する。

### 2.3 Upsert → Get: 保存したレコードが取得できる
- **操作**: `SummaryRecord` を `Upsert` し、同じ `cacheKey` で `Get` する。
- **期待**: 保存した内容と一致するレコードが返る。

### 2.4 Upsert (上書き): 同一cacheKeyで更新される
- **操作**: 同一 `cacheKey` で2回 `Upsert` する（2回目は `summary_text` を変更）。
- **期待**: `Get` で2回目の内容が返る。`updated_at` が更新されている。

### 2.5 Get: 存在しないcacheKeyでnilが返る
- **操作**: 存在しない `cacheKey` で `Get` する。
- **期待**: `nil, nil`（エラーなし、レコードなし）が返る。

### 2.6 GetByRecordID: レコードIDと種別で検索できる
- **操作**: dialogue型とquest型のレコードを保存し、`GetByRecordID` で種別指定検索する。
- **期待**: 指定した種別のレコードのみが返る。

### 2.7 GetByRecordID: 存在しない組み合わせでnilが返る
- **操作**: 存在しない `recordID` で `GetByRecordID` する。
- **期待**: `nil, nil` が返る。

### 2.8 Close: DB接続が正常にクローズされる
- **操作**: `Close()` を呼び出す。
- **期待**: エラーなし。以降のDB操作はエラーとなる。

### 2.9 異なるソースファイルで独立したDBが作成される
- **操作**: `NewSQLiteSummaryStore(tempDir, "Skyrim.esm")` と `NewSQLiteSummaryStore(tempDir, "Dawnguard.esm")` を作成し、それぞれにレコードを保存する。
- **期待**: 各DBに保存したレコードが互いに干渉しない。

---

## 3. SummaryGeneratorImpl テスト

### 3.1 会話要約: キャッシュミス時にLLM呼び出しと保存が行われる
- **モック**: `SummaryStore.Get` → nil, `LLMClient.Complete` → 成功レスポンス
- **入力**: 1件の `DialogueGroupInput`（lines=3行）
- **期待**:
  - `LLMClient.Complete` が1回呼ばれる。
  - `SummaryStore.Upsert` が1回呼ばれる。
  - 結果の `SummaryResult.CacheHit` が `false`。

### 3.2 会話要約: キャッシュヒット時にLLM呼び出しがスキップされる
- **モック**: `SummaryStore.Get` → 既存レコード
- **入力**: 1件の `DialogueGroupInput`
- **期待**:
  - `LLMClient.Complete` が呼ばれない。
  - 結果の `SummaryResult.CacheHit` が `true`。
  - 結果の `SummaryText` がキャッシュの値と一致。

### 3.3 会話要約: 空の入力行はスキップされる
- **入力**: `DialogueGroupInput` の `Lines` が空。
- **期待**: `LLMClient.Complete` が呼ばれず、`SummaryText` が空文字。

### 3.4 クエスト要約: ステージがIndex昇順でソートされる
- **モック**: `SummaryStore.Get` → nil, `LLMClient.Complete` → 成功
- **入力**: `QuestInput` の `StageTexts` が `["stage3", "stage1", "stage2"]`（意図的に順序を崩す）
- **期待**: `LLMClient.Complete` に渡されるプロンプトで `stage1`, `stage2`, `stage3` の順になっている。
- **注**: `QuestInput.StageTexts` は呼び出し側がIndex昇順でソート済みで渡す前提。本テストはGeneratorImpl内部でのプロンプト構築を検証する。

### 3.5 クエスト要約: 空のステージはスキップされる
- **入力**: `QuestInput` の `StageTexts` が空。
- **期待**: `LLMClient.Complete` が呼ばれず、`SummaryText` が空文字。

### 3.6 LLMエラー時に空文字が返りプロセスが継続する
- **モック**: `LLMClient.Complete` → エラーレスポンス
- **入力**: 1件の `DialogueGroupInput`
- **期待**: `SummaryResult.SummaryText` が空文字。エラーがログ出力される。パニックしない。

### 3.7 複数入力の並列処理と進捗通知
- **モック**: `SummaryStore.Get` → nil, `LLMClient.Complete` → 成功（遅延あり）
- **入力**: 5件の `DialogueGroupInput`
- **期待**:
  - 5件の `SummaryResult` が返る。
  - `progress` コールバックが5回呼ばれる（done=1〜5）。

### 3.8 プロンプト構築: 会話要約のシステムプロンプトが正しい
- **検証**: `buildDialoguePrompt` が生成するユーザープロンプトが `"Summarize the following conversation:\n- line1\n- line2"` 形式であること。

### 3.9 プロンプト構築: クエスト要約のシステムプロンプトが正しい
- **検証**: `buildQuestPrompt` が生成するユーザープロンプトが `"Summarize the overall quest based on the descriptions of these quest stages:\n- stage1\n- stage2"` 形式であること。

---

## 4. 統合テスト（インメモリSQLite）

### 4.1 End-to-End: 会話要約の生成→キャッシュ→再取得
- **セットアップ**: インメモリSQLite + モックLLMClient
- **操作**:
  1. `GenerateDialogueSummaries` を呼び出す（キャッシュミス → LLM生成 → 保存）。
  2. 同一入力で再度 `GenerateDialogueSummaries` を呼び出す。
- **期待**: 2回目はキャッシュヒットし、LLMが呼ばれない。

### 4.2 End-to-End: 入力変更時にキャッシュが無効化される
- **操作**:
  1. `GenerateDialogueSummaries` を呼び出す（lines=`["A", "B"]`）。
  2. 同一GroupIDで異なるlines（`["A", "B", "C"]`）で再度呼び出す。
- **期待**: 2回目はキャッシュミスとなり、LLMが再度呼ばれる。新しい要約で上書きされる。

### 4.3 End-to-End: 異なるソースファイルのキャッシュが独立している
- **セットアップ**: 2つの異なるソースファイル用 `SQLiteSummaryStore` を作成
- **操作**:
  1. ソースファイルAのStoreにレコードを保存する。
  2. ソースファイルBのStoreで同じ `cacheKey` を検索する。
- **期待**: ソースファイルBではキャッシュミスとなる（DBが独立しているため）。
