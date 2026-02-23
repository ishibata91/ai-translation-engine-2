## Why

2-Pass翻訳のPass 2において、文脈構築（Lore）と翻訳実行（Translator）が分離していることで、中間データ構造（`TranslationRequest`）の受け渡しがプロセスマネージャー（オーケストレーター）を介して行われており、アーキテクチャが無駄に複雑化しています。これを垂直スライス原則に基づき、`terminology` スライスと同様の「データ入力からLLMリクエスト生成まで」を完結させる構造に統合し、保守性と直感性を向上させます。

## What Changes

- **スライスの統合**: `lore` スライスの文脈構築ロジック（会話ツリー解析、話者プロファイリング等）を `translator` スライスに合流させ、単一の `Pass2Translator` スライスとして再定義します。
- **データフローの簡略化**: `ProcessManager` からの呼び出しフローを `GameData -> TranslationRequest -> llm.Request` から、直接 `GameData -> llm.Request` を生成する **2フェーズモデル (Propose/Save)** へ変更します。
- **中間DTOの廃止**: スライス間を跨いでいた `TranslationRequest` 構造体を廃止、またはスライス内部のプライベートなデータ構造に隠蔽します。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `translator`: `lore` の文脈構築ロジックを統合し、ゲームデータ入力から直接LLMプロンプトを構築する自律的な垂直スライスに拡張。
- `lore`: 独立した垂直スライスとしての定義を解除し、`translator` スライス内の文脈構築ロジック（Context Engine）として再配置。

## Impact

- **`ProcessManager`**: `lore` と `translator` を個別に呼び出していたオーケストレーションロジックが大幅に簡素化されます。
- **パッケージ構造**: `pkg/lore` と `pkg/translator` の統合、または `pkg/translator` 内への `lore` ロジックの移動が発生します。
- **インターフェース**: `Lore` インターフェースが廃止され、統合された `Translator` インターフェースがゲームデータを受け取る形式に変更されます。
