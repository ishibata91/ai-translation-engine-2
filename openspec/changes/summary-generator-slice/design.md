# Design: SummaryGeneratorSlice

## Context

現行Python版ツール（v1.x）では、会話・クエスト要約をサイドカーJSONファイル（`summary_cache.py`）にキャッシュしているが、並列処理時の書き込み競合・検索性の低さ・Mod単位での管理の困難さが課題となっている。
Go v2移行に伴い、永続化先をSQLiteへ変更し、完全自律Vertical Sliceとして`SummaryGeneratorSlice`を実装する。

## Goals / Non-Goals

**Goals:**
- 会話（`DialogueGroup`）およびクエスト（`Quest`）の要約をLLMで生成し、ソースプラグイン単位のSQLiteファイルにキャッシュする独立スライスを実装する。
- Consumer-Driven Contractsの原則に基づき、本スライス専用の入力DTO（`SummaryGeneratorInput`）を `contract.go` 内に定義し、外部パッケージのモデルに依存しない。
- SHA-256ハッシュに基づくキャッシュキーでLLM呼び出しの重複を防止する。
- クエスト要約はステージを時系列に累積処理し、「これまでのあらすじ」を段階的にビルドする。

**Non-Goals:**
- Pass 2での実際の翻訳処理（要約の活用はPass 2スライスの責務）。
- 会話ツリーの構造解析（ContextEngineSliceの責務）。
- グローバルなデータモデル（`pkg/domain`等）の作成や共有。

## Decisions

1. **ソースプラグイン単位SQLiteキャッシュファイル**
   - **Rationale**: `{plugin_name}_summary_cache.db` としてMod単位に分割することで、Mod別の削除・再生成・配布が容易になる。他スライスのテーブルとの同居を避けることでスライスの自律性を保つ。

2. **Consumer-Driven Contracts DTO**
   - **Rationale**: `SummaryGeneratorInput`（`DialogueItems` / `QuestItems`）を本スライス内に定義し、LoaderSliceやContextEngineSliceのDTO変更による破壊的影響を受けない。変換はオーケストレーター層の責務とする。

3. **SHA-256キャッシュキー方式**
   - **Rationale**: `{record_id}|{sha256_hash}` 形式のキャッシュキーにより、入力テキストの変更を即座に検知し不整合を防ぐ。レコードIDとハッシュの複合検索でインデックスが効き高速なHIT判定が可能。

4. **カプセル化された永続化層（`SummaryStore`）**
   - **Rationale**: 外部からは `*sql.DB` のみを受け取り、`summaries` テーブルのDDL・UPSERT・SELECTはスライス内に閉じ込める。さらに `Init()` 内で `PRAGMA journal_mode=WAL` と `PRAGMA busy_timeout=5000` を自己発行することで、並行書き込みロック耐性をスライスが自律的に確保する。上位コンポーネント（ProcessManager等）にSQLite設定の責務を持たせない。

5. **2フェーズモデル（Propose/Save）の採用**
   - **Rationale**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成（ProposeJobs）」と「結果保存（SaveResults）」の2段階に分離する。スライス自身はLLM Clientを直接呼び出さず、通信制御はインフラ層（JobQueue/ProcessManager）へ委譲することで、ネットワーク分断やバッチAPI待機に対する堅牢性を確保する。
   - **Phase 1 (Propose)**: 入力データを解析し、キャッシュヒット判定を行う。既訳がない場合はLLMプロンプトを生成してリクエスト群として返す。既訳がある場合は即時結果として返す。
   - **Phase 2 (Save)**: JobQueue等を通じて取得されたLLMのレスポンス群を受け取り、パースしてSQLiteキャッシュに永続化する。

6. **構造化ログ基盤への適合（slog-otel）**
   - **Rationale**: `refactoring_strategy.md` セクション6「テスト戦略」・セクション7「構造化ログ基盤」に準拠し、全Contractメソッドの入口・出口で `slog.DebugContext(ctx, ...)` を出力。TraceIDを全ログに自動付与する。

## Risks / Trade-offs

- **LLM呼び出しコストと処理時間**
  - **Risk**: 多数の`DialogueGroup`・`Quest`に対して逐次LLMアクセスを行うとAPIコストと処理時間が増大する。
  - **Mitigation**: SHA-256キャッシュにより再処理コストをほぼゼロに抑え、Goroutineによる並列処理で実行時間を削減する。

- **SQLiteの並行書き込みロック**
  - **Risk**: 複数Goroutineから同時にUPSERTを行うと `database is locked` エラーが発生し得る。
  - **Mitigation**: `SummaryStore` の初期化処理（`Init()`）内で `PRAGMA journal_mode=WAL` および `PRAGMA busy_timeout=5000` を自己発行する。上位コンポーネント（ProcessManager等）への依存は排除し、スライスが自律的にロック耐性を確保する。

- **バッチAIモードでの非効率（1件ずつ発行問題）**
  - **Risk**: 1件ずつ `Generate` を呼ぶ設計では、バッチAPIモード時に「N個の独立したバッチジョブ」が個別に作成・ポーリングされ、バッチの意味がなくなる。
  - **Mitigation**: 2フェーズモデルによる一括提案で解決。スライスはMISSプロンプトを全件まとめて `ProposeJobs` の戻り値として返し、上位の `ProcessManager` が `LLMClient.GenerateBulk` に渡す。これによりバッチジョブの集約が自然に行われる。

- **クエスト要約の累積処理の順序依存**
  - **Risk**: ステージを昇順に処理する必要があり、並列化と順序保証が競合する。
  - **Mitigation**: クエスト単位では逐次処理（ステージIndex昇順）を維持しつつ、異なるクエスト間は並列実行する設計とする。
