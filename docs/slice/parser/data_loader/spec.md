# データローダー
> Responsible for Phase 1 Data Foundation and domain models.

## Purpose
TBD: Parsing JSON files into domain models.

## Requirements

### Requirement: Structured domain models by context
The existing `models` package MUST retain the `ExtractedData` and related domain constructs, but split them from a single file into context-specific structural files (e.g. dialogue, quest, entity).

#### Scenario: Referencing domain models
- **WHEN** a component imports `github.com/ishibata91/ai-translation-engine-2/pkg/domain/models`
- **THEN** it can cleanly access models such as `NPC` or `Quest` without depending on a massive, monolithic file, while backward compatibility of the package path is maintained.

---

## ログ出力・テスト共通規約

> 本スライスは `standard_test_spec.md` と `log-guide.md` に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（`standard_test_spec.md` 参照）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（`log-guide.md` 参照）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（`log-guide.md` 参照）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（`log-guide.md` 参照）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（`log-guide.md` の AI デバッグ運用参照）でログと仕様書をAIに渡し修正させる。
