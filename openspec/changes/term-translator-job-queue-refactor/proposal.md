# Proposal: Term Translator Job Queue Refactor

## Problem & Motivation
現在実装済みの `TermTranslatorSlice` は、自らの内部で `LLMClient` を呼び出し、結果を待ってから `Mod用語DB` へ保存する「同期的なPull型」のアーキテクチャになっています。
しかし、新しい「JobQueueインフラ」と「ProcessManager」のVSA統合アーキテクチャにおいては、スライスは「プロンプト生成」と「結果保存」の2フェーズに分かれ、通信や進捗管理の泥臭い処理は完全にインフラ層へ委譲されるべきです。
現在のままでは、TermTranslatorがバッチAPIモードの長時間待機に対応できず、またProcessManagerのエコシステム（UI連携など）にも乗ることができません。

## Proposal
既存の `TermTranslatorSlice` の設計と実装を破棄（または大幅リファクタリング）し、新しい「Slice-Owned Callback パターン」に適合させます。
具体的には、従来の「処理をすべて一貫して行う太い関数」を破棄し、ProcessManagerから呼ばれる「Phase 1: プロンプト（Request）生成関数」と、「Phase 2: LLM結果（Response）を受け取ってSQLiteに保存するコールバック関数」の2つにドメインロジックを分離します。

## Capabilities

### Modified Capabilities
- `TermTranslator`: 既存の単一実行フローを破棄し、プロンプト生成と結果保存の2フェーズモデルへ再設計。

## Impact
- `pkg/term_translator` 内部の実装の抜本的変更。
- 外部API（インターフェース）が `GenerateTranslations` のような形から、`PreparePrompts` と `SaveResults` のペアに変更される。
