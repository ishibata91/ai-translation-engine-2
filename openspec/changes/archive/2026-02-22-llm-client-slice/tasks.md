## 1. 基盤と型定義 (Foundation & Types)

- [x] 1.1 `pkg/llm_client` パッケージを作成し、`LLMClient` インターフェース、`TranslationRequest` などのDTOを定義する。
- [x] 1.2 `ConfigStoreSlice` 側から渡されるプロバイダ設定（Gemini, Local, xAI 等）を表すモデルを定義する。

## 2. コアロジック実装 (Core Implementation)

- [x] 2.1 `gemini`, `local`, `xai` などの各具象クライアント構造体を実装する。
- [x] 2.2 各プロバイダにおいて、`HTTP 429` や `50x` エラー発生時に行われる、インフラ起因のバックオフ・リトライ処理を組み込む。
- [x] 2.3 設定値に基づき適切な具象クライアントを初期化して返すファクトリ `LLMManager` を実装する。

## 3. ロギングとテレメトリ統合 (Logging & Telemetry)

- [x] 3.1 各具象クライアントのリクエスト前後に、OpenTelemetryのTraceIDを利用した `slog.DebugContext` を仕込む。
- [x] 3.2 構造化ログ機能が利用可能かを確認するための `slog` ヘルパーを調整または導入する。

## 4. テスト (Testing)

- [x] 4.1 各具象クライアントのインフラリトライロジック（`HTTP 429` などのモック）がバックオフ等に従って正しく再試行されるかのパラメタライズドテストを作成する。
- [x] 4.2 `LLMManager` によるプロバイダ動的解決のテストを行う。
