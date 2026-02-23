## ADDED Requirements

### Requirement: SummaryGeneratorSliceの独立した初期化
本スライスは独自の入力DTO（`SummaryGeneratorInput`）を `contract.go` 内に定義し、他スライスのデータ構造に依存してはならない。

#### Scenario: オーケストレーターが入力DTOをSummaryGeneratorSliceに渡す
- **WHEN** オーケストレーターが LoaderSlice / ContextEngineSlice の出力から会話・クエストデータを抽出し、`SummaryGeneratorInput` にマッピングして渡す
- **THEN** SummaryGeneratorSlice が外部パッケージの型に依存することなく初期化され、要約生成の準備が整う

### Requirement: 会話要約の生成
本スライスは `DialogueGroup` 単位で、`PreviousID` チェーンを含む会話フローを収集し、LLMにより500文字以内の英語1文として要約を生成しなければならない。

#### Scenario: 会話フローを収集してLLMで要約を生成する
- **WHEN** `DialogueGroup` 内に1件以上の会話行があり、LLMが正常に応答する
- **THEN** 「誰が誰に何について話しているか」を明確にした500文字以内の英語1文要約が返却される

#### Scenario: 会話行が0件の場合はスキップする
- **WHEN** `DialogueGroup` 内の会話行が0件である
- **THEN** LLMは呼び出されず、空のサマリとして処理がスキップされる

### Requirement: クエスト要約の累積生成
本スライスは `Quest.Stages` を `Index` 昇順に処理し、各ステージ処理時点での過去ステージ記述を累積してLLMへ送信し、「これまでのあらすじ」を生成しなければならない。

#### Scenario: 複数ステージのクエスト要約を累積生成する
- **WHEN** `Quest` が2件以上のステージを持ち、LLMが正常に応答する
- **THEN** 各ステージの処理で過去すべてのステージ記述が入力として使われ、累積的なあらすじが生成される

#### Scenario: ステージ記述が0件の場合はスキップする
- **WHEN** `Quest.Stages` が空である
- **THEN** LLMは呼び出されず、処理がスキップされる

### Requirement: SHA-256キャッシュキーによるキャッシュヒット判定
本スライスは `{record_id}|{sha256_hash}` 形式のキャッシュキーでSQLiteを検索し、ヒット時はLLM呼び出しをスキップしなければならない。

#### Scenario: キャッシュヒット時にLLM呼び出しをスキップする
- **WHEN** 同一の `record_id` と入力テキストに対して要約生成が再度リクエストされる
- **THEN** SQLiteのキャッシュから要約テキストが返却され、LLMへのリクエストは生成されない

#### Scenario: キャッシュミス時に要約を生成してキャッシュに保存する
- **WHEN** 該当するキャッシュキーがSQLiteに存在しない
- **THEN** LLMによって要約が生成され、`summaries` テーブルにUPSERTされる

### Requirement: ソースプラグイン単位SQLiteへの要約永続化（SummaryStore）
本スライスは `summaries` テーブルのDDL・UPSERT・SELECTをスライス内部にカプセル化し、外部からは `*sql.DB` のみを受け取る。

#### Scenario: 初回起動時にsummariesテーブルを自動作成する
- **WHEN** `SummaryStore` が初期化される
- **THEN** 接続先SQLiteに `summaries` テーブルが存在しない場合、`CREATE TABLE IF NOT EXISTS` により自動作成される

#### Scenario: 要約をUPSERTで保存する
- **WHEN** 要約生成が成功した後にストアへの保存が呼ばれる
- **THEN** `summaries` テーブルに対して UPSERT（同一 `cache_key` の場合は上書き）が実行される

### Requirement: 並列要約生成
本スライスは複数の `DialogueGroup` および `Quest` の要約を Goroutine で並列処理しなければならない。並列度は Config で設定可能（デフォルト: 10）とする。ただし、同一クエスト内のステージは逐次処理（Index昇順）を維持する。

#### Scenario: 複数DialogueGroupを並列処理する
- **WHEN** 複数の `DialogueGroup` が入力として渡される
- **THEN** 設定された並列度でGoroutineが起動し、各グループの要約が並列に生成される

#### Scenario: クエストのステージはIndex昇順に逐次処理する
- **WHEN** 同一 `Quest` の複数ステージを処理する
- **THEN** ステージは `Index` 昇順に逐次処理され、要約が累積的にビルドされる

### Requirement: 構造化ログおよびTraceID伝播
本スライスは `refactoring_strategy.md` セクション6・7に準拠し、全 Contract メソッドの入口・出口で `slog.DebugContext(ctx, ...)` を出力し、TraceID を全ログに伝播しなければならない。

#### Scenario: 要約生成開始時のEntryログ出力
- **WHEN** 要約生成処理が開始される
- **THEN** `trace_id`・`span_id`・入力件数を含む DEBUG レベルの Entry ログが JSON 形式で出力される

#### Scenario: 要約生成完了時のExitログ出力
- **WHEN** 要約生成処理が完了する
- **THEN** 生成件数・スキップ件数・経過時間を含む DEBUG レベルの Exit ログが JSON 形式で出力される
