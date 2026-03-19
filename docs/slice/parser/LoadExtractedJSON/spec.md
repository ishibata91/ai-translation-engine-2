# 抽出済み JSON 読み込み

## Requirements

### Requirement: JSON Loading

抽出されたJSONファイルを読み込み、`ExtractedData` 構造体にマッピングする。

#### Scenario: Successful Load
- **WHEN** 有効なパスと有効なJSONファイルが渡されたとき
- **THEN** エラーなし (`nil`) を返す
- **AND** 返却される `*ExtractedData` 構造体の各フィールド (Quests, NPC, Items 等) にデータが格納されている

#### Scenario: File Not Found
- **WHEN** 存在しないファイルパスが渡されたとき
- **THEN** `os.ErrNotExist` ラップしたエラーを返す
- **AND** `nil` を返す

#### Scenario: Invalid JSON Syntax
- **WHEN** JSON構文が壊れているファイルが渡されたとき
- **THEN** JSONデコードエラーを返す
- **AND** `nil` を返す

#### Scenario: Partial Data Loading
- **WHEN** `quests` フィールドのみが存在するJSONが渡されたとき
- **THEN** エラーなしを返す
- **AND** `ExtractedData.Quests` にデータが格納されている
- **AND** 他のフィールド (NPCs, Items 等) は空またはゼロ値である

---

## ログ出力・テスト共通規約

> 本スライスは `standard_test_spec.md` と `log-guide.md` に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（`standard_test_spec.md` 参照）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（`log-guide.md` 参照）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（`log-guide.md` 参照）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（`log-guide.md` 参照）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（`log-guide.md` の AI デバッグ運用参照）でログと仕様書をAIに渡し修正させる。
