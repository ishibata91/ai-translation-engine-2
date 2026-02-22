## ADDED Requirements

### Requirement: プロバイダ非依存のLLMクライアントインターフェース
LLM呼び出しを行う各スライスは、個別のプロバイダ（Gemini, Localなど）の実装ではなく、統一された `LLMClient` インターフェースを使用して翻訳・テキスト生成を実行しなければならない (MUST)。

#### Scenario: 翻訳リクエストの実行
- **WHEN** TermTranslatorSliceが `LLMClient` の `Translate` メソッドを呼び出した時
- **THEN** 内部で設定された適切なLLMプロバイダにリクエストが送信され、エラーなく結果の文字列が返却されること

### Requirement: 実行時のLLMプロバイダ動的解決
`LLMManager`（ファクトリ）は、渡された設定情報（プロバイダ種別、APIキーなど）に基づいて、対応する具象 `LLMClient` インスタンスを動的に生成し返却しなければならない (MUST)。

#### Scenario: Geminiクライアントの生成
- **WHEN** ユーザーが設定で「Gemini」を選択し、その設定情報が `LLMManager` に渡された時
- **THEN** `LLMManager` は Gemini API に接続する実装の `LLMClient` を返すこと

#### Scenario: xAIクライアントの生成
- **WHEN** ユーザーが設定で「xAI」を選択し、その設定情報が `LLMManager` に渡された時
- **THEN** `LLMManager` は xAI API に接続する実装の `LLMClient` を返すこと

### Requirement: インフラ起因エラーの自律的リトライ
LLMClientは、プロバイダ固有の通信エラー（例: HTTP 429 Too Many Requests や 503 Service Unavailable）に対して、呼び出し元のスライスにエラーを返す前に、内部でExponential Backoff等を用いた適切な再試行を行わなければならない (MUST)。

#### Scenario: レートリミット時の自動リトライ
- **WHEN** LLMAPIから `HTTP 429 Too Many Requests` が返却された時
- **THEN** Clientは直ちにエラーを返さず、指定されたバックオフ戦略に従い待機した後にリクエストを自動で再試行すること

#### Scenario: ビジネス要件の抽出・リトライの各スライス委譲
- **WHEN** LLMAPIが正常に `HTTP 200 OK` を返したが、内容が各スライスの期待するフォーマット（例: `TL: |テキスト|`）に合致しない時
- **THEN** Clientは文字列をそのまま返し、フォーマット違反の判定および再試行の要否（ビジネスロジック起因のリトライ）は呼び出し元のスライスが行うこと

### Requirement: 構造化ログ（slog）とTraceIDの統合
すべてのLLMリクエストとレスポンス、およびエラーは、OpenTelemetryのTraceIDを含んだ構造化ログ (`slog`) として記録されなければならない (MUST)。

#### Scenario: APIリクエスト時のログ出力
- **WHEN** `LLMClient` が外部のAPIにリクエストを送信する前後
- **THEN** `slog.DebugContext` 等を用いて、実行時間や使用トークン数、TraceIDを含むJSONログが出力されること
