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
