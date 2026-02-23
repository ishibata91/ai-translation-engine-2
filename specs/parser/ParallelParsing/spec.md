# ParallelParsing Spec

## Requirements

### Requirement: Parallel Mapping & Normalization

JSONのロード自体はシングルスレッドで行い、その後の構造体へのマッピング（Unmarshal）と正規化処理を並列化する。

#### Scenario: Two-Phase Loading
- **WHEN** `LoadExtractedJSON` が呼び出されたとき
- **THEN** まず `map[string]json.RawMessage` としてトップレベルのキーをデコードする (Serial)
- **AND** その後、`quests`, `dialogue_groups` などの重いフィールドの Unmarshal と正規化を個別の Goroutine で開始する (Parallel)

#### Scenario: Normalization Overhead
- **WHEN** 正規化処理（ID抽出やバリデーション）が走るとき
- **THEN** 各Goroutine内で並列に実行され、メインスレッドのブロックを防ぐ

#### Scenario: Error Aggregation
- **WHEN** 並列処理中に1つ以上のマッピングエラーが発生したとき
- **THEN** エラーを収集し、処理完了後にまとめて（または最初のクリティカルエラーを）返す

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
