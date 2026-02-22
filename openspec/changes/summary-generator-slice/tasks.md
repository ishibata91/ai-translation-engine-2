## 1. コア構造と DTO

- [ ] 1.1 `pkg/summary_gen` パッケージを作成し、`contract.go` を定義する
- [ ] 1.2 グローバルな `pkg/domain` モデルに依存しない `SummaryGeneratorInput` DTO（`DialogueItems`・`QuestItems`）を定義する
- [ ] 1.3 スライス内にコアインターフェース（`SummaryGenerator`・`SummaryStore`）を定義する
- [ ] 1.4 Wire 依存性注入のための DI プロバイダ（`provider.go`）を設定する

## 2. キャッシュキー生成とヒット判定

- [ ] 2.1 `{record_id}|{sha256_hash}` 形式のキャッシュキーを生成するロジックを実装する（Go標準 `crypto/sha256`）
- [ ] 2.2 SQLiteに対してキャッシュキーで検索し、ヒット時はLLM呼び出しをスキップするキャッシュヒット判定を実装する

## 3. データベースと SummaryStore

- [ ] 3.1 `*sql.DB` をDIで受け取る `SummaryStore` を実装する。`Init()` 内で `PRAGMA journal_mode=WAL` と `PRAGMA busy_timeout=5000` を自己発行し、上位コンポーネントへの依存なしにロック耐性を確保する
- [ ] 3.2 `summaries` テーブルのDDL（`CREATE TABLE IF NOT EXISTS`）をスライス内にカプセル化し、初期化時に自動作成する
- [ ] 3.3 `idx_summaries_cache_key`・`idx_summaries_record_type` インデックスを作成する
- [ ] 3.4 `UpsertSummary` メソッドを実装する（同一 `cache_key` の場合は上書き）
- [ ] 3.5 `GetSummary(cacheKey string)` メソッドを実装する

## 4. 会話要約の生成

- [ ] 4.1 `DialogueGroup` 単位で `PreviousID` チェーンを辿り会話フローを構築するロジックを実装する
- [ ] 4.2 システムプロンプト・ユーザープロンプトを構築し、`LLMClient` へリクエストするメソッドを実装する
- [ ] 4.3 入力行が0件の場合はスキップする条件分岐を追加する

## 5. クエスト要約の累積生成

- [ ] 5.1 `Quest.Stages` を `Index` 昇順でソートするロジックを実装する
- [ ] 5.2 ステージを逐次処理し、過去ステージ記述を累積してLLMへ送信する累積要約ロジックを実装する
- [ ] 5.3 ステージ記述が0件の場合はスキップする条件分岐を追加する

## 6. バルクリクエスト処理

- [ ] 6.1 全対象レコードのキャッシュチェックをGoroutineで並列実行し、HITリストとMISSリストに仕分けるロジックを実装する
- [ ] 6.2 MISSプロンプトを `[]GenerateRequest` としてまとめ、`LLMClient.GenerateBulk(ctx, reqs)` を一括呼び出しするロジックを実装する
- [ ] 6.3 クエスト要約の累積処理は同一クエスト内でステージ逐次・クエスト間は並列でキャッシュチェックを行い、各クエストのMISSプロンプトをバルクリストに集約する
- [ ] 6.4 `context.Context` のキャンセルを全Goroutineに伝播させる

## 7. 構造化ログ

- [ ] 7.1 全 Contract メソッドの入口・出口に `slog.DebugContext(ctx, ...)` による Entry/Exit ログを追加する
- [ ] 7.2 `trace_id`・`span_id`・処理件数・経過時間をログフィールドに含める

## 8. テスト

- [ ] 8.1 インメモリSQLite（`:memory:`）を利用したスライスレベルのパラメタライズドテスト（`generator_test.go`）を作成する
- [ ] 8.2 `GenerateBulk` をモック化し、実際のAPI呼び出しなしにパイプライン全体を検証する
- [ ] 8.3 キャッシュヒット時のLLMスキップ動作（MISSリストに含まれないこと）を検証する
- [ ] 8.4 バルクリクエストで渡されるプロンプト件数がキャッシュMISS件数と一致することを検証する
- [ ] 8.5 クエスト要約の累積処理（ステージ逐次）の動作を検証する
- [ ] 8.6 入力行0件時のスキップ動作を検証する

