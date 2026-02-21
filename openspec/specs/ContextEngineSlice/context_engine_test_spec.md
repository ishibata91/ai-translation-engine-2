# コンテキストエンジン テスト設計

## テスト方針
- 全テストはモック（ToneResolver, PersonaLookup, TermLookup, SummaryLookup）を使用し、外部依存なしで実行可能とする。
- SQLite系Lookup実装のテストにはインメモリSQLite (`:memory:`) を使用する。

---

## 1. ToneResolver テスト

### 1.1 種族マッピングが正しく適用される
- **入力**: race=`"KhajiitRace"`, voiceType=`""`, sex=`"Male"`
- **期待**: カジート固有の口調指示が含まれる。

### 1.2 ボイスタイプマッピングが正しく適用される
- **入力**: race=`""`, voiceType=`"MaleCommander"`, sex=`"Male"`
- **期待**: MaleCommander固有の口調指示が含まれる。

### 1.3 種族+ボイスタイプの複合指示が生成される
- **入力**: race=`"OrcRace"`, voiceType=`"MaleCoward"`, sex=`"Male"`
- **期待**: 種族指示とボイスタイプ指示が改行で結合される。

### 1.4 ボイスタイプ不明時に性別フォールバックが適用される（Female）
- **入力**: race=`""`, voiceType=`""`, sex=`"Female"`
- **期待**: `"標準的な女性の口調"` を含む指示が返る。

### 1.5 ボイスタイプ不明時に性別フォールバックが適用される（Male）
- **入力**: race=`""`, voiceType=`""`, sex=`"Male"`
- **期待**: `"標準的な男性の口調"` を含む指示が返る。

### 1.6 全属性が空の場合のフォールバック
- **入力**: race=`""`, voiceType=`""`, sex=`""`
- **期待**: 標準的な男性の口調（デフォルト）が返る。パニックしない。

---

## 2. PersonaLookup テスト（SQLite実装）

### 2.1 存在するSpeakerIDでペルソナが取得できる
- **セットアップ**: `npc_personas` テーブルにレコードを挿入。
- **操作**: `FindBySpeakerID("NPC001")` を呼び出す。
- **期待**: 保存した `persona_text` が返る。

### 2.2 存在しないSpeakerIDでnilが返る
- **操作**: `FindBySpeakerID("NONEXISTENT")` を呼び出す。
- **期待**: `nil, nil`（エラーなし、レコードなし）が返る。

### 2.3 SpeakerIDが空文字の場合
- **操作**: `FindBySpeakerID("")` を呼び出す。
- **期待**: `nil, nil` が返る。パニックしない。

---

## 3. TermLookup テスト（SQLite実装）

### 3.1 完全一致で参照用語と強制翻訳が返る
- **セットアップ**: 辞書DBに `"Iron Sword"` → `"鉄の剣"` を登録。
- **操作**: `Search("Iron Sword")` を呼び出す。
- **期待**: `forcedTranslation` = `"鉄の剣"`, `referenceTerms` に `"Iron Sword"` が含まれる。

### 3.2 キーワード完全一致で参照用語が返る
- **セットアップ**: 辞書DBに `"Iron"` → `"鉄"`, `"Sword"` → `"剣"` を登録。
- **操作**: `Search("Iron Sword of Fire")` を呼び出す。
- **期待**: `forcedTranslation` = nil, `referenceTerms` に `"Iron"` と `"Sword"` が含まれる。

### 3.3 貪欲最長一致で短い候補が抑制される
- **セットアップ**: 辞書DBに `"Broken Tower Redoubt"` → `"壊れた塔の砦"`, `"Broken"` → `"壊れた"` を登録。
- **操作**: `Search("Broken Tower Redoubt is nearby")` を呼び出す。
- **期待**: `referenceTerms` に `"Broken Tower Redoubt"` のみが含まれ、`"Broken"` 単独は含まれない。

### 3.4 NPC名の部分一致検索
- **セットアップ**: NPC FTS5テーブルに `"Jon Battle-Born"` → `"ジョン・バトルボーン"` を登録。
- **操作**: `Search("Battle-Born family")` を呼び出す。
- **期待**: `referenceTerms` に `"Jon Battle-Born"` が含まれる。

### 3.5 ステミングによる形態変化対応
- **セットアップ**: 辞書DBに `"Sword"` → `"剣"` を登録。
- **操作**: `Search("Daedric Swords")` を呼び出す。
- **期待**: `referenceTerms` に `"Sword"` が含まれる（`Swords` → stem `sword` → マッチ）。

### 3.6 所有格の除去とステミング
- **セットアップ**: 辞書DBに `"Auriel"` → `"アウリエル"` を登録。
- **操作**: `Search("Auriel's Bow")` を呼び出す。
- **期待**: `referenceTerms` に `"Auriel"` が含まれる。

### 3.7 辞書DBが空の場合
- **操作**: 空の辞書DBに対して `Search("anything")` を呼び出す。
- **期待**: 空の `referenceTerms`, `forcedTranslation` = nil。エラーなし。

### 3.8 複数DB横断検索
- **セットアップ**: DB1に `"Iron"` → `"鉄"`, DB2に `"Steel"` → `"鋼"` を登録。
- **操作**: `Search("Iron and Steel")` を呼び出す。
- **期待**: `referenceTerms` に `"Iron"` と `"Steel"` の両方が含まれる。

---

## 4. SummaryLookup テスト（SQLite実装）

### 4.1 会話要約が取得できる
- **セットアップ**: `summaries` テーブルに `record_id="DIAL001"`, `summary_type="dialogue"` のレコードを挿入。
- **操作**: `FindDialogueSummary("DIAL001")` を呼び出す。
- **期待**: 保存した `summary_text` が返る。

### 4.2 クエスト要約が取得できる
- **セットアップ**: `summaries` テーブルに `record_id="QUST001"`, `summary_type="quest"` のレコードを挿入。
- **操作**: `FindQuestSummary("QUST001")` を呼び出す。
- **期待**: 保存した `summary_text` が返る。

### 4.3 存在しないIDでnilが返る
- **操作**: `FindDialogueSummary("NONEXISTENT")` を呼び出す。
- **期待**: `nil, nil` が返る。

### 4.4 種別の不一致でnilが返る
- **セットアップ**: `summaries` テーブルに `record_id="DIAL001"`, `summary_type="dialogue"` のレコードを挿入。
- **操作**: `FindQuestSummary("DIAL001")` を呼び出す。
- **期待**: `nil, nil` が返る（種別が不一致）。

---

## 5. ContextEngineImpl テスト

### 5.1 会話リクエスト: INFO NAM1が正しく生成される
- **モック**: ToneResolver → 固定指示, PersonaLookup → nil, TermLookup → 空, SummaryLookup → nil
- **入力**: 1件のDialogueGroup（1件のDialogueResponse、SpeakerID付き）
- **期待**: INFO NAM1のTranslationRequestが1件生成される。Speaker情報が設定されている。

### 5.2 会話リクエスト: Previous Lineが正しく追跡される
- **入力**: 1件のDialogueGroup（PlayerText="Hello", 2件のResponse）
- **期待**:
  - 1件目のINFO NAM1の `PreviousLine` が `"Hello"`（PlayerText）。
  - 2件目のINFO NAM1の `PreviousLine` が1件目のResponse.Text。

### 5.3 会話リクエスト: DIAL FULLが生成される
- **入力**: DialogueGroup（PlayerText="Choose wisely"）
- **期待**: DIAL FULLのTranslationRequestが生成される。RecordType=`"DIAL FULL"`。

### 5.4 会話リクエスト: PlayerTextがnilの場合DIAL FULLがスキップされる
- **入力**: DialogueGroup（PlayerText=nil）
- **期待**: DIAL FULLのリクエストが生成されない。

### 5.5 会話リクエスト: INFO RNAMが生成される
- **入力**: DialogueResponse（MenuDisplayText="Yes, I'll help"）
- **期待**: INFO RNAMのTranslationRequestが生成される。

### 5.6 会話リクエスト: 日本語テキストがスキップされる
- **入力**: DialogueResponse（Text="こんにちは"）
- **期待**: INFO NAM1のリクエストが生成されない。

### 5.7 会話リクエスト: 空テキストがスキップされる
- **入力**: DialogueResponse（Text=""）
- **期待**: INFO NAM1のリクエストが生成されない。

### 5.8 会話リクエスト: SpeakerIDがnilの場合Speakerがnilになる
- **入力**: DialogueResponse（SpeakerID=nil）
- **期待**: TranslationRequest.Context.Speaker が nil。

### 5.9 会話リクエスト: ペルソナが存在する場合SpeakerProfileに設定される
- **モック**: PersonaLookup → `"勇敢な戦士。粗野だが義理堅い。"`
- **入力**: DialogueResponse（SpeakerID="NPC001"）
- **期待**: SpeakerProfile.PersonaText が設定されている。

### 5.10 会話リクエスト: 会話要約がコンテキストに含まれる
- **モック**: SummaryLookup.FindDialogueSummary → `"Ulfric orders march on Whiterun"`
- **入力**: 1件のDialogueGroup
- **期待**: TranslationRequest.Context.DialogueSummary が設定されている。

### 5.11 トピック名: TopicText優先
- **入力**: DialogueResponse（TopicText="Greetings"）, DialogueGroup（NAM1="Topic1"）
- **期待**: TopicName = `"Greetings"`。

### 5.12 トピック名: TopicTextがnilの場合NAM1にフォールバック
- **入力**: DialogueResponse（TopicText=nil）, DialogueGroup（NAM1="Topic1"）
- **期待**: TopicName = `"Topic1"`。

### 5.13 トピック名: 100文字超過時に切り詰められる
- **入力**: TopicText = 120文字の文字列
- **期待**: TopicName が100文字以内（97文字 + `"..."`）。

### 5.14 クエストリクエスト: QUST FULLが生成される
- **入力**: Quest（Name="The Golden Claw"）
- **期待**: QUST FULLのTranslationRequestが生成される。

### 5.15 クエストリクエスト: QUST CNAMがステージごとに生成される
- **入力**: Quest（Stages=3件）
- **期待**: QUST CNAMのTranslationRequestが3件生成される。各リクエストにIndexが設定されている。

### 5.16 クエストリクエスト: QUST NNAMが目標ごとに生成される
- **入力**: Quest（Objectives=2件）
- **期待**: QUST NNAMのTranslationRequestが2件生成される。

### 5.17 クエストリクエスト: クエスト要約がCNAM/NNAMに含まれる
- **モック**: SummaryLookup.FindQuestSummary → `"Retrieve the Dragonstone"`
- **入力**: Quest（Stages=1件）
- **期待**: QUST CNAMのContext.QuestSummary が設定されている。

### 5.18 クエストリクエスト: QUST FULLにはクエスト要約が含まれない
- **モック**: SummaryLookup.FindQuestSummary → `"summary"`
- **入力**: Quest（Name="Test Quest"）
- **期待**: QUST FULLのContext.QuestSummary が nil。

### 5.19 アイテムリクエスト: {Type} DESCが生成される
- **入力**: Item（Description="A sharp blade", TypeHint="Weapon"）
- **期待**: TranslationRequestが生成される。Context.ItemTypeHint = `"Weapon"`。

### 5.20 アイテムリクエスト: BOOK DESCが生成される
- **入力**: Item（Text="<p>Long book text...</p>", TypeHint="Book"）
- **期待**: BOOK DESCのTranslationRequestが生成される。

### 5.21 強制翻訳: 辞書完全一致時にForcedTranslationが設定される
- **モック**: TermLookup.Search → forcedTranslation=`"鉄の剣"`
- **入力**: Item（Description="Iron Sword"）
- **期待**: TranslationRequest.ForcedTranslation = `"鉄の剣"`。

### 5.22 参照用語: ReferenceTermsが設定される
- **モック**: TermLookup.Search → referenceTerms=[{OriginalEN:"Iron", OriginalJA:"鉄"}]
- **入力**: 任意のレコード
- **期待**: TranslationRequest.ReferenceTerms に用語が含まれる。

### 5.23 SourcePlugin/SourceFileが全リクエストに設定される
- **入力**: config.SourceFile=`"MyMod_Export.json"`, レコードのSource=`"MyMod.esp"`
- **期待**: 全TranslationRequestに SourcePlugin と SourceFile が設定されている。

---

## 6. 統合テスト（インメモリSQLite）

### 6.1 End-to-End: 全レコードタイプの翻訳リクエスト生成
- **セットアップ**: インメモリSQLite + モック各Lookup
- **入力**: DialogueGroup×1, Quest×1, Item×1, Magic×1, Message×1 を含む ExtractedData
- **操作**: `BuildTranslationRequests(data, config)` を呼び出す。
- **期待**: 各レコードタイプに対応するTranslationRequestが生成される。

### 6.2 End-to-End: 空のExtractedDataで空リストが返る
- **入力**: 全フィールドが空の ExtractedData
- **期待**: 空の `[]TranslationRequest` が返る。エラーなし。

### 6.3 End-to-End: Lookup系エラー時にプロセスが継続する
- **モック**: TermLookup.Search → エラー
- **入力**: 1件のDialogueGroup
- **期待**: エラーがログ出力されるが、パニックせず処理が継続する。参照用語は空で設定される。
