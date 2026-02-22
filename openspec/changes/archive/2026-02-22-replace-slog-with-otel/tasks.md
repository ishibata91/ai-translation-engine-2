# Tasks: Replace slog with slog-otel

## 1. 準備と依存関係 (Setup & Dependencies)
- [x] 1.1 `github.com/samber/slog-otel` を `go.mod` に追加する
- [x] 1.2 `go mod tidy` を実行して依存関係を整理する

## 2. インフラストラクチャの更新 (Infrastructure)
- [x] 2.1 `pkg/infrastructure` （または main）でのロガー初期化処理において、`slog-otel` ハンドラでラップするように修正する
- [x] 2.2 `google/wire` を使用している場合、プロバイダ設定を更新し `wire` を再実行する

## 3. 各スライスの移行 (Slice Migration)
- [x] 3.1 `pkg/term_translator` の全てのログ出力を `slog.*Context(ctx, ...)` に置換する
- [x] 3.2 `pkg/dictionary_builder` の全てのログ出力を `slog.*Context(ctx, ...)` に置換する
- [x] 3.3 `pkg/config_store` の全てのログ出力を `slog.*Context(ctx, ...)` に置換する
- [x] 3.4 その他 `pkg/` 以下の残りの `slog` 利用箇所を同様に修正する
- [x] 3.5 テストコード内のロガー利用箇所も必要に応じて修正する

## 4. 検証 (Verification)
- [x] 4.1 全てのパッケージが正常にビルドできることを確認する
- [x] 4.2 実行時のログ（JSON 形式）に `trace_id` と `span_id` が期待通り出力されていることを目視またはテストで確認する
