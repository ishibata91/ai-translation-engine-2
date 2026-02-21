# Loader Slice Architecture
> Interface-First AIDD: Loader Vertical Slice

## Purpose
TBD: Provision of dependency injection for the loader module and encapsulation of internal data handling.

## Requirements

### Requirement: Loader slice DI Provider
The `pkg/loader` module MUST provide a dependency injection provider function via Google Wire, hiding its internal implementation details.

#### Scenario: DI Initialization
- **WHEN** the application starts and initializes its components
- **THEN** it resolves `contract.Loader` through the module's Wire provider without directly instantiating internal structs.

### Requirement: Internal process encapsulation
The sequential load and parallel decoding steps MUST be encapsulated within the `contract.Loader` implementation and not exposed to the Process Manager.

#### Scenario: Orchestrating the load process
- **WHEN** the Process Manager invokes `LoadExtractedJSON`
- **THEN** the slice internally coordinates file reading, decoding, and structuring, returning only the final `ExtractedData` or an error.

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
