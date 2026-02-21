# EncodingDetection Spec

## Requirements

### Requirement: Automatic Encoding Detection

入力ファイルのエンコーディングを自動判別し、UTF-8に変換して読み込む。

#### Scenario: UTF-8 (Standard)
- **WHEN** UTF-8 でエンコードされたJSONファイルが渡されたとき
- **THEN** 正常にデコードできる
- **AND** 文字化けが発生しない

#### Scenario: UTF-8 with BOM
- **WHEN** BOM付きUTF-8 (UTF-8-SIG) でエンコードされたファイルが渡されたとき
- **THEN** BOMを無視して正常にデコードできる

#### Scenario: Shift-JIS (Japanese Legacy)
- **WHEN** Shift-JIS (CP932) でエンコードされたファイルが渡されたとき
- **THEN** 正常にデコードできる
- **AND** 日本語文字が正しく保持される

#### Scenario: CP1252 (European Legacy)
- **WHEN** CP1252 (Latin-1) でエンコードされたファイルが渡されたとき
- **THEN** 正常にデコードできる
- **AND** 特殊文字（アクセント記号など）が正しく保持される

#### Scenario: Unknown Encoding
- **WHEN** 上記いずれのエンコーディングでもデコードに失敗した場合
- **THEN** エラーを返す

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
