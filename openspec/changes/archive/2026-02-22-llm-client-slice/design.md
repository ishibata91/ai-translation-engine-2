# Design: LLM Client Slice

## Context
現在 `translate_with_local_ai.py` から Go/React へのVSAアーキテクチャ移行が進められており、各Vertical Slice（TermTranslatorSlice, ContextEngineSlice 等）がLLMに依存する設計となっている。
各スライスにLLMクライアント機能（プロンプト構築、HTTP通信、エラーハンドリング）をそれぞれ実装すると、「DRYな責務の重複」が発生するだけでなく、全ての箇所で複数のLLM（Gemini, Local, xAI）の切り替えロジックを持つことになり、保守性が著しく低下する。

このため、LLM通信の抽象化とプロバイダ動的解決の責務を担う共通基盤（`LLMClient` インターフェース、およびファクトリとしての `LLMManager`）を設計する。

## Goals / Non-Goals
**Goals:**
* Goの `interface` を用いた、LLMプロバイダに依存しない共通契約 (`LLMClient`) の定義。
* `ConfigStoreSlice` の設定値に基づき、適切な具象実装（Gemini, Local, xAI 等）を生成するファクトリ機構の提供。
* 統一された出力フォーマット（`TL: |...|`）の抽出・フォールバックロジックのカプセル化。
* 構造化ログ（slog + OpenTelemetryのTraceID）を用いたデバッグ容易性の確保。

**Non-Goals:**
* 本スライス自体がドメインロジック（例: ゲーム特有の会話ツリー解析）を持つこと。
* 他のスライスの内部設計（DTO等）を変更すること。（本モジュールは各種スライスから呼び出されるUtility層または独立スライスとして機能する）

## Decisions

### 1. 共通インターフェース (`LLMClient`)
```go
type GenerateRequest struct {
    SystemPrompt string
    UserPrompt   string
    Temperature  float32
    // その他基本的なLLMパラメータ
}

type LLMClient interface {
    // LLMにテキスト生成リクエストを行い、生の文字列を返す
    GenerateText(ctx context.Context, req GenerateRequest) (string, error)
}
```
**Rationale**: 各スライスが実装を意識せず呼べるインターフェース。旧案で存在した翻訳フォーマット抽出は削除し、各スライスの自律性を保護する。

### 2. 動的プロバイダファクトリ (`LLMManager`)
`ConfigStoreSlice` から取得した設定値（LLMプロバイダ、APIキー、ベースURL等）を引数に取り、対応するクライアント実装を返すファクトリメソッドを用意する。
**Rationale**: VSAの原則に従い、各スライスは「その時点のユーザー設定」を `ConfigStore` から取得し、それを `LLMManager` に渡すことで実行時にプロバイダを切り替えるため。設定ファイルのパース等はファクトリ側で行わず、呼び出し元のContextから与えられるようにする。

### 3. エラーハンドリングの分離 (インフラ vs ビジネス)
LLMClientの実装は、`HTTP 429 Too Many Requests` や `HTTP 502/503/504` 等のプロバイダ側のインフラエラーに対するExponential Backoffによるリトライのみを隠蔽して処理する。
**Rationale**: LLM特有のインフラ制御を各スライスがWETに実装する事を避けるため。一方「TLフォーマットが見つからない」「意図しない形式の返答だったため再生成したい」などのビジネス要件上のリトライや、特定のタイムアウト制御は、それぞれが独立したドメイン知識となるためスライス側に委譲し、WETの実装を許容（推奨）する。

## Risks / Trade-offs

* **[Risk] パイプフォーマット抽出等のWETな実装**: 各スライス内で「TLフォーマットの抽出漏れ」などバグの温床になり得る。
  * **Mitigation**: AI駆動開発（AIDD）において、ユーティリティを明示的に指定し忘れるよりは、プロンプト指示によってそれぞれの要件に合わせた正規表現ロジック等を各スライス内で生成させる方が、コードの自律性と保守性が高い（VSAの原則）。
* **[Risk] プロバイダ固有の拡張機能の利用困難化**: 抽象化することで、特定のLLM（例: Geminiの関数呼び出し等）に依存した固有機能が使いづらくなる。
  * **Mitigation**: 今回のユースケース（テキスト翻訳・要約）に特化したインターフェースに絞り、高度な機能が必要になった場合は別途インターフェースを拡張する。
* **[Risk] 設定のリアルタイム更新とインスタンスのライフサイクル**: 実行中にユーザーがLLM設定を変更した場合の反映。
  * **Mitigation**: 各スライスは処理ごと（またはバッチごと）に毎回 `LLMManager` に最新の設定を渡して再生成するか、設定変更イベントをリッスンする設計とする。

## Migration Plan
1. `pkg/llm_client` 等にインターフェースとモジュールを作成し、各種ダミー実装（またはローカル実装）を追加する。
2. その後、`TermTranslatorSlice` などの既存実装を修正し、この `LLMClient` に依存するようにリファクタリングする。
