# 本文翻訳 テスト設計

## テスト方針
- 全テストはモック（LLMClient, PromptBuilder, TagProcessor, ResultWriter, BookChunker, TranslationVerifier）を使用し、外部依存なしで実行可能とする。
- ResultWriterのテストには一時ディレクトリを使用する。

---

## 1. TagProcessor テスト

### 1.1 Preprocess: HTMLタグがプレースホルダーに置換される
- **入力**: `"Hello <font color='red'>World</font>"`
- **期待**: processedText=`"Hello [TAG_1]World[TAG_2]"`, tagMap に2エントリ。

### 1.2 Preprocess: aliasタグが置換される
- **入力**: `"Talk to <alias=QuestGiver>"`
- **期待**: aliasタグがプレースホルダーに置換される。

### 1.3 Preprocess: タグなしテキストはそのまま返る
- **入力**: `"No tags here"`
- **期待**: processedText=`"No tags here"`, tagMap が空。

### 1.4 Postprocess: プレースホルダーが元のタグに復元される
- **入力**: text=`"こんにちは [TAG_1]世界[TAG_2]"`, tagMap=`{[TAG_1]:<font>, [TAG_2]:</font>}`
- **期待**: `"こんにちは <font>世界</font>"`

### 1.5 Postprocess: 未使用プレースホルダーが残っている場合
- **入力**: tagMapに存在しないプレースホルダーがテキストに含まれる
- **期待**: 未使用プレースホルダーはそのまま残る（エラーにしない）。

### 1.6 Validate: 原文にないタグが検出される
- **入力**: translatedText=`"<b>太字</b>"`, tagMap にbタグなし
- **期待**: `TagHallucinationError` が返る。

### 1.7 Validate: 正常なタグ復元でエラーなし
- **入力**: translatedText=`"<font>テスト</font>"`, tagMap に対応エントリあり
- **期待**: nil（エラーなし）。

### 1.8 Preprocess: 複数の同一タグが個別プレースホルダーになる
- **入力**: `"<br>line1<br>line2<br>"`
- **期待**: 3つの個別プレースホルダーが生成される。

---

## 2. PromptBuilder テスト

### 2.1 会話系: INFO NAM1のプロンプトが正しく構築される
- **入力**: TranslationRequest（RecordType="INFO NAM1", Speaker付き）
- **期待**: systemPromptに話者情報（名前・種族・口調）が含まれる。

### 2.2 会話系: プレイヤー発言（INFO RNAM）のプロンプト
- **入力**: TranslationRequest（RecordType="INFO RNAM", Speaker=nil）
- **期待**: systemPromptに "Player" が話者として含まれる。

### 2.3 会話系: ペルソナテキストがプロンプトに含まれる
- **入力**: TranslationRequest（Speaker.PersonaText="勇敢な戦士"）
- **期待**: systemPromptにペルソナテキストが含まれる。

### 2.4 クエスト系: QUST CNAMのプロンプトが正しく構築される
- **入力**: TranslationRequest（RecordType="QUST CNAM", QuestName="The Golden Claw", QuestSummary="Retrieve..."）
- **期待**: systemPromptにクエスト名と要約が含まれる。

### 2.5 アイテム系: DESC のプロンプトにTypeHintが含まれる
- **入力**: TranslationRequest（RecordType="WEAP DESC", ItemTypeHint="Weapon"）
- **期待**: systemPromptにアイテム種別が含まれる。

### 2.6 書籍系: BOOK DESCのプロンプトに書籍固有指示が含まれる
- **入力**: TranslationRequest（RecordType="BOOK DESC"）
- **期待**: systemPromptに文体維持・段落構造保持の指示が含まれる。

### 2.7 汎用: 未知のレコードタイプでもエラーにならない
- **入力**: TranslationRequest（RecordType="UNKNOWN TYPE"）
- **期待**: 基本的なsystemPromptが返る。エラーなし。

### 2.8 参照用語がプロンプトに含まれる
- **入力**: TranslationRequest（ReferenceTerms=[{OriginalEN:"Iron", OriginalJA:"鉄"}]）
- **期待**: userPromptまたはsystemPromptに参照用語リストが含まれる。

### 2.9 Config Storeからカスタムテンプレートが取得される
- **モック**: Config Store → カスタムテンプレート
- **期待**: カスタムテンプレートが使用される。

### 2.10 Config Storeにテンプレートがない場合デフォルトにフォールバック
- **モック**: Config Store → 未設定
- **期待**: デフォルトテンプレートが使用される。

---

## 3. Translator テスト（単一リクエスト）

### 3.1 正常系: LLM翻訳が成功する
- **モック**: PromptBuilder → 成功, TagProcessor → パススルー, LLMClient → 成功レスポンス
- **入力**: TranslationRequest（SourceText="Hello"）
- **期待**: TranslationResult{Status:"success", TranslatedText:"こんにちは"}

### 3.2 強制翻訳: ForcedTranslationが設定されている場合LLMをスキップ
- **入力**: TranslationRequest（ForcedTranslation="鉄の剣"）
- **期待**: LLMClient.Complete が呼ばれない。Status="success", TranslatedText="鉄の剣"。

### 3.3 スキップ: 空テキスト
- **入力**: TranslationRequest（SourceText=""）
- **期待**: Status="skipped"。LLMClient.Complete が呼ばれない。

### 3.4 スキップ: 空白のみのテキスト
- **入力**: TranslationRequest（SourceText="   "）
- **期待**: Status="skipped"。

### 3.5 HTMLタグ: 前処理→翻訳→後処理の一連フロー
- **モック**: TagProcessor.Preprocess → プレースホルダー置換, LLMClient → 翻訳成功, TagProcessor.Postprocess → タグ復元
- **入力**: TranslationRequest（SourceText=`"<font>Hello</font>"`）
- **期待**: 翻訳結果にHTMLタグが正しく復元されている。

### 3.6 タグハルシネーション: リトライされる
- **モック**: TagProcessor.Validate → 1回目TagHallucinationError, 2回目nil
- **期待**: LLMClient.Complete が2回呼ばれる。最終的にStatus="success"。

### 3.7 リトライ上限: max_retries到達で失敗
- **モック**: LLMClient.Complete → 常にエラー
- **期待**: Status="failed"。ErrorMessage が設定されている。LLMClient.Complete が max_retries 回呼ばれる。

### 3.8 リトライ不可エラー: context.Canceledで即座に失敗
- **モック**: LLMClient.Complete → context.Canceled
- **期待**: リトライせず即座にStatus="failed"。

### 3.9 max_tokensの動的計算
- **入力**: SourceText = 100トークン相当のテキスト
- **期待**: LLMRequestのMaxTokensが `max(100 * 2.5, 100)` = 250 に設定される。

### 3.10 max_tokensの最小値保証
- **入力**: SourceText = 非常に短いテキスト（5トークン）
- **期待**: MaxTokens が 100（最小値）に設定される。

---

## 4. BookChunker テスト

### 4.1 短いテキストは分割されない
- **入力**: text="Short text", maxTokens=1500
- **期待**: 1チャンクのみ返る。

### 4.2 段落単位で分割される
- **入力**: text=`"<p>Para1</p><p>Para2</p><p>Para3</p>"`, maxTokens=各段落1つ分
- **期待**: 3チャンクに分割される。各チャンクがHTMLタグ構造を維持。

### 4.3 長大な単一段落が文単位で分割される
- **入力**: 段落タグなしの長文テキスト, maxTokens=小さい値
- **期待**: 文単位で分割される。

### 4.4 空テキストで空配列が返る
- **入力**: text=""
- **期待**: 空の[]string が返る。

---

## 5. TranslationVerifier テスト

### 5.1 正常な翻訳でエラーなし
- **入力**: source="Hello", translated="こんにちは", tagMap=空
- **期待**: nil。

### 5.2 空の翻訳結果でエラー
- **入力**: source="Hello", translated="", tagMap=空
- **期待**: エラーが返る。

### 5.3 原文コピーでエラー
- **入力**: source="Hello", translated="Hello", tagMap=空
- **期待**: エラーが返る（翻訳されていない可能性）。

### 5.4 タグ不整合でエラー
- **入力**: source=`"<font>Hello</font>"`, translated=`"こんにちは"`, tagMap にfontタグあり
- **期待**: タグ不整合エラーが返る。

---

## 6. ResultWriter テスト

### 6.1 Write: 翻訳結果がファイルに保存される
- **セットアップ**: 一時ディレクトリ
- **操作**: Write(result) を呼び出す。
- **期待**: 出力ディレクトリにJSONファイルが作成される。

### 6.2 Write: 同一プラグインの結果が追記される
- **操作**: 同一SourcePluginで2回 Write を呼び出す。
- **期待**: JSONファイルに2件のレコードが含まれる。

### 6.3 Write: 異なるプラグインで別ファイルに保存される
- **操作**: 異なるSourcePluginで Write を呼び出す。
- **期待**: プラグインごとに別のJSONファイルが作成される。

### 6.4 Flush: バッファ内の全結果がファイルに書き出される
- **操作**: 複数回 Write 後に Flush を呼び出す。
- **期待**: 全結果がファイルに反映されている。

### 6.5 出力フォーマット: JSON構造が正しい
- **操作**: Write → Flush 後にファイルを読み込む。
- **期待**: `[{"form_id":..., "editor_id":..., "type":..., "original":..., "string":...}]` 形式。

### 6.6 並行Write: データ競合が発生しない
- **操作**: 複数Goroutineから同時に Write を呼び出す。
- **期待**: `go test -race` でデータ競合が検出されない。

---

## 7. BatchTranslator テスト

### 7.1 正常系: 全リクエストが翻訳される
- **モック**: Translator → 全成功, ResultWriter → 成功
- **入力**: 3件のTranslationRequest
- **期待**: 3件のTranslationResult（全てstatus="success"）。

### 7.2 Resume: 既存翻訳済みリクエストがスキップされる
- **モック**: ResumeLoader → 1件のキャッシュ済み結果
- **入力**: 3件のTranslationRequest（うち1件がキャッシュ済み）
- **期待**: Translator.Translate が2回のみ呼ばれる。1件はstatus="cached"。

### 7.3 リクエストキー生成: 正しいキー形式
- **入力**: SourcePlugin="Skyrim.esm", ID="0x001234", RecordType="INFO NAM1", Index=nil
- **期待**: キー = `"Skyrim.esm|0x001234|INFO NAM1"`

### 7.4 リクエストキー生成: Index付き
- **入力**: SourcePlugin="Skyrim.esm", ID="0x005678", RecordType="QUST CNAM", Index=10
- **期待**: キー = `"Skyrim.esm|0x005678|QUST CNAM|10"`

### 7.5 並列実行: MaxWorkers制御
- **モック**: Translator.Translate → 遅延あり（100ms）
- **入力**: 10件のリクエスト, MaxWorkers=2
- **期待**: 同時実行数が2を超えない（実行時間から推定）。

### 7.6 コンテキストキャンセル: 処理が中断される
- **操作**: TranslateBatch開始後にctxをキャンセルする。
- **期待**: 未処理のリクエストがスキップされ、エラーが返る。

### 7.7 進捗通知: コールバックが正しく呼ばれる
- **入力**: 5件のリクエスト
- **期待**: progressコールバックが5回呼ばれる（done=1〜5）。

### 7.8 翻訳失敗: 一部失敗でもプロセスが継続する
- **モック**: Translator.Translate → 2件目でエラー
- **入力**: 3件のリクエスト
- **期待**: 3件全てのTranslationResultが返る。2件目はstatus="failed"。

### 7.9 書籍翻訳: BOOK DESCがチャンク分割翻訳される
- **モック**: BookChunker → 3チャンク, Translator → 各チャンク成功
- **入力**: TranslationRequest（RecordType="BOOK DESC", 長文テキスト）
- **期待**: 翻訳結果が結合されたテキストとなる。

---

## 8. 統合テスト

### 8.1 End-to-End: リクエスト→翻訳→保存の一連フロー
- **セットアップ**: モックLLMClient + 一時ディレクトリ
- **操作**: TranslateBatch を呼び出す。
- **期待**: 出力ディレクトリにJSONファイルが生成され、翻訳結果が正しく保存されている。

### 8.2 End-to-End: Resume動作の検証
- **操作**:
  1. TranslateBatch を呼び出す（3件翻訳）。
  2. 同一リクエストで再度 TranslateBatch を呼び出す。
- **期待**: 2回目はLLMが呼ばれず、全件status="cached"。

### 8.3 End-to-End: 混合ステータスの結果
- **モック**: LLMClient → 1件成功, 1件失敗, 1件は強制翻訳
- **入力**: 3件のリクエスト
- **期待**: 結果に success, failed, success（forced）が混在し、全てファイルに保存される。
