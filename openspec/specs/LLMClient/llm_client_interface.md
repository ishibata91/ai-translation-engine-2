# LLMクライアントインターフェース設計 (LLM Client Interface Design)

## 1. 概要 (Overview)
本ドキュメントは、全スライスで利用可能な共通インフラストラクチャとしてのLLMクライアントインターフェースを定義する。
Interface-First AIDD 原則に基づき、具体的なLLMプロバイダー（OpenAI, Gemini, xAI, Local/GGUFなど）の実装から独立した、抽象的な契約（Contract）を定義することを目的とする。

## 2. 設計方針 (Design Principles)
- **プロトコル指向**: 特定の構造体ではなく、インターフェースを介したアクセスを基本とする。
- **並行処理のネイティブサポート**: GoのGoroutinesとChannelsを活用した効率的なリクエスト処理。
- **バッチ処理とストリーミング**: 単発リクエスト、ストリーミングレスポンス、および大規模な非同期バッチAPI（xAI Batch API等）を統一的に扱う。
- **構造化出力 (Structured Output)**: JSONスキーマに基づいた型安全なレスポンス取得。

## 3. 主要コンポーネント (Components)
- `LLMClient`: LLMへのリクエスト送信を担当するコアインターフェース。
- `LLMProvider`: 特定のプロバイダー（Gemini, OpenAI等）の実装。
- `LLMManager`: 利用可能なプロバイダーを管理し、設定に基づいたインスタンス提供を行う。
- `BatchHandler`: 非同期バッチAPIのジョブ管理と結果取得を抽象化する。

## 4. クラス図とシーケンス図
詳細な図は別ファイルを参照。
- [クラス図 (Class Diagram)](llm_client_class_diagram.md)
- [シーケンス図 (Sequence Diagram)](llm_client_sequence_diagram.md)

## 5. テスト設計 (Test Strategy)
詳細なテスト仕様は別ファイルを参照。
- [テスト設計 (Test Spec)](llm_client_test_spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
