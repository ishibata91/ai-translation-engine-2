# Proposal: LLM Client Slice

## Motivation
現在のシステムでは、翻訳処理や用語抽出など複数のユースケース（TermTranslatorSlice, ContextEngineSlice, DictionaryBuilderSliceなど）でAIによるテキスト生成（LLM）が必要となる。
`refactoring_strategy.md` および `requirements.md` に従い、各スライスが特定のLLMプロバイダ（Gemini, Local, xAI等）の具象実装に直接依存せず、シームレスに切り替えられるようにする必要がある。また、翻訳エンジンとしての一貫性や、フォーマット抽出（`TL: |にほんご|`）の責務を共通化し、各スライスの自律性を保ちつつもLLM呼び出しの複雑さを隠蔽する基盤が求められている。

## Goals
* **統合されたLLM Clientインターフェースの提供**: 呼び出し元のスライス（Context EngineやTerm Translator等）に対し、プロバイダ非依存の統一された通信用インターフェースを提供する。
* **動的プロバイダ解決**: `ConfigStoreSlice` から受け取る設定情報（例: `namespace: llm`）に基づいて、ローカル、Gemini、xAI、バッチAPIモードなどを実行時に動的に選択・初期化するファクトリ（`LLMManager`）を実装する。
* **インフラ起因のエラー制御のカプセル化**: APIのレートリミット（HTTP 429）や通信の一時的切断に対するバックオフ制御のみを `LLMClient` に閉じ込め、ビジネス要件に直結する抽出ロジックやタイムアウト制御は各スライスへ委譲する。
* **構造化ログ（slog + OpenTelemetryのTraceID）連携**: 各種リクエスト・レスポンスの統合的な可視化とデバッグ容易性の確保。

## Capabilities
### New Capabilities
- `llm-client`: 統一インターフェース、プロバイダ動的解決、インフラエラー制御機能を持つLLMクライアント基盤。

### Modified Capabilities
- 既存の各スライス（TermTranslatorSlice等）は、この `llm-client` が提供する機能を利用するように統合される（本Changeのスコープ外だが影響範囲にあたる）。各スライスは自身のドメイン要件に応じた独自のフォーマット抽出と自律的なビジネスリトライの責務を引き継ぐ。

## Impact
- **アーキテクチャ**: 各ドメインスライスからLLM通信の詳細（APIキー管理、HTTPリクエスト構築など）が排除され、Contractへの依存のみとなる。
- **UI/UX設定**: ユーザーがReact UIのモーダル等から一元的にLLMを切り替え・設定可能となる（本Changeはバックエンド基盤の提供）。
- **テスト**: 外部API呼び出しをモックしやすくなり、独立したパラメタライズドテストの信頼性が向上する。
