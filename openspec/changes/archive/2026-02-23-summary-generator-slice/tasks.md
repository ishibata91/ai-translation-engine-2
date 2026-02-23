## 1. コア構造と DTO

- [x] 1.1 `pkg/summary_generator` パッケージを作成し、`contract.go` を定義する
- [x] 1.2 グローバルな `pkg/domain` モデルに依存しない `SummaryGeneratorInput` DTO（`DialogueItems`・`QuestItems`）を定義する
- [x] 1.3 スライス内にコアインターフェース（`SummaryGenerator`・`SummaryStore`）を定義する
- [x] 1.4 Wire 依存性注入のための DI プロバイダ（`provider.go`）を設定する

## 2. キャッシュキー生成とヒット判定

- [x] 2.1 `{record_id}|{sha256_hash}` 形式のキャッシュキーを生成するロジックを実装する（Go標準 `crypto/sha256`）
- [x] 2.2 SQLiteに対してキャッシュキーで検索し、ヒット時はLLM呼び出しをスキップするキャッシュヒット判定を実装する

## 3. 永続化層（SummaryStore）の実装

- [x] 3.1 `*sql.DB` をDIで受け取る `SummaryStore` を実装する
- [x] 3.2 `Init()` 内で `PRAGMA` 設定（WAL/busy_timeout）を行い、自律的にロック耐性を確保する
- [x] 3.3 `summaries` テーブルのDDLカプセル化と初期化時の自動作成を実装する
- [x] 3.4 検索用インデックス（`cache_key`, `record_type`）を作成する
- [x] 3.5 低レベルなデータ操作メソッド（`Get`, `Upsert`, `GetByRecordID`）を実装する

## 4. プロンプト構築ロジック

- [x] 4.1 `DialogueGroup` 単位で `PreviousID` チェーンを辿り会話フローを構築するロジックを実装する
- [x] 4.2 会話要約用のシステムプロンプト・ユーザープロンプトを構築するロジックを実装する
- [x] 4.3 クエストの `StageTexts` を `Index` 昇順でソートし、累積的なプロンプトを構築するロジックを実装する
- [x] 4.4 入力行が0件の場合はジョブ提案から除外する条件分岐を追加する

## 5. 2フェーズモデル・コントラクト実装

- [x] 5.1 `ProposeJobs` メソッドの並列化実装
  - [x] `SummaryStore.Get` を用いてキャッシュ判定を **並列実行** し、HIT/MISSを仕分ける（要 `errgroup`）
  - [x] 設定された並列度（デフォルト 10）で Goroutine 数を制御する
  - [x] MISS分についてプロンプトを構築し、`[]llm_client.Request` を生成する
  - [x] HIT分を `PreCalculatedResults` としてまとめ、`ProposeOutput` を返す
- [x] 5.2 `SaveResults` メソッドを実装する
  - [x] `[]llm_client.Response` から要約を抽出し、`SummaryStore.Upsert` で永続化する
- [x] 5.3 `GetSummary` メソッドを実装し、`SummaryStore.GetByRecordID` を呼び出して Pass 2 へ要約を返す
- [x] 5.4 `context.Context` のキャンセルを内部の並列処理に伝播させる

## 6. 構造化ログと OpenTelemetry

- [x] 6.1 `ProposeJobs`・`SaveResults`・`GetSummary` の入口・出口に `slog.DebugContext(ctx, ...)` を追加する
- [x] 6.2 `trace_id`・処理件数・経過時間をログフィールドに含める

## 7. テスト

- [x] 7.1 インメモリSQLite（`:memory:`）を利用したスライスレベルのパラメタライズドテスト（`generator_test.go`）を作成する
- [x] 7.2 `ProposeJobs` のテスト: キャッシュMISS時に正しいプロンプトが生成されること、HIT時にDBから値が引かれることを検証する
- [x] 7.3 `SaveResults` のテスト: レスポンスが正しくDBに保存されること、バリデーション失敗時にスキップされることを検証する
- [x] 7.4 クエスト要約の累積プロンプト構築ロジックを検証する
- [x] 7.5 入力行0件時のスキップ動作を検証する

## 8. 修正・リファクタリング（TODO）

- [x] 8.1 `ProposeJobs` の処理を `errgroup` を用いた並列処理にリファクタリングする
- [x] 8.2 `SummaryGeneratorConfig` を導入し、並列度を外部から設定可能にする
- [x] 8.3 テストに Quest 側の 0件スキップケースを追加する

