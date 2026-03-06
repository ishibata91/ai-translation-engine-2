# LoadExtractedJSON Spec

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

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
