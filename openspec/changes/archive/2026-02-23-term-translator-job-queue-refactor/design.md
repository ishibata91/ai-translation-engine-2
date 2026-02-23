# Design: Term Translator Job Queue Refactor

## Context
「ProcessManager」と「汎用JobQueue」の確立により、各ドメインスライスはLLMとの直接的な通信責務（goroutineでの待機、バッチAPIでのポーリング等）から完全に解放されます。既存の `TermTranslatorSlice` は自らLLMClientを叩いていましたが、この責務を取り上げ、より純粋なVSA（ロジックと永続化のみ）へと純化させます。

## Goals / Non-Goals

**Goals:**
- `TermTranslatorSlice` を「入力からLLMリクエスト（`llm_client.Request`）配列を生成する関数」と「LLMレスポンス（`llm_client.Response`）配列を受け取ってSQLiteにUPSERTする関数」に分割する。
- 内部で保持していたチャネル制御やLLMエラーハンドリングコードを破棄する。
- 従来通り、内部のSQLite（`Mod用語DB`）への最終的な保存や、FTSインデックスの更新といった「ドメインの永続化責務」は完全に維持する。

**Non-Goals:**
- 翻訳プロンプトの内容そのもの（LLMにどういうコンテキストを渡すか）の変更。プロンプト構築ロジックは既存のものを流用する。

## Decisions

1. **フェーズの完全分離 (Two-Phase Function Contract)**:
   - **[Principle] 既存実装の破棄**: リクエスト生成（プロンプト構築）ロジックを除き、以前の同期実行型関数（`GenerateTranslations` 等）はすべて破棄します。中途半端な修正ではなく、新しい VSA コントラクトに基づいた再実装を行います。
   - `PreparePrompts(ctx, input TermTranslatorInput) ([]llm_client.Request, error)`: 辞書引き等のコンテキスト構築を行い、純粋なプロンプトだけを返す。
   - `SaveResults(ctx, input TermTranslatorInput, results []llm_client.Response) error`: 返ってきた結果をパース（`TL: |にほんご|` のパイプ抽出等）し、自身のSQLiteへ保存する。
   （※元のInput情報が必要な場合、ProcessManagerが仲介して保持・再注入する設計とする）

2. **Errorハンドリングの単純化**:
   - `SaveResults` は、渡されたレスポンスのうち `Success == false` なものを単にスキップ（またはエラーログを残す）するだけでよくなり、通信エラー起因のリトライ制御などは一切記述しなくて済むようになります。

3. **リファクタリング戦略の適用 (Compliance with Refactoring Strategy)**:
   - **Interface-First AIDD**: `pkg/term_translator/contract.go` でインターフェースと独自の DTO（`TermTranslatorInput` 等）を定義し、スライス間の疎結合を維持する。
   - **Method-Level SRP**: 巨大な `PreparePrompts` ロジックは、同一ファイル内のプライベートメソッド（`buildSinglePrompt`, `lookupDictionary` 等）に分割し、認知負荷を下げる。
   - **構造化ログ基盤**: `slog` を使用し、Entry/Exit ログに TraceID を含めることで、JobQueue と ProcessManager を跨いだ追跡を可能にする。

## Risks / Trade-offs

- **[Risk] 入力と出力のマッチング**:
  Phase 1で出力したRequest配列と、Phase 2で入力されるResponse配列の順序が完全に一致している保証（または紐付けキー）が必要になります。
  → **Mitigation**: JobQueueの実装仕様として「渡されたRequestの配列順序と、返されるResponseの配列順序は必ず1:1で一致する」という制約をインフラ側に設け、スライス側は単純に配列インデックス単位で突合できる設計とします。
