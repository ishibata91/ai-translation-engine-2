# Design: Persona Gen Job Queue Refactor

## Context
ProcessManagerと汎用JobQueueの導入に合わせて、各ドメインスライスから「LLMの完了を待機する」責務を排除します。
`PersonaGenSlice` は数千人のNPCデータを処理する可能性があり、その通信処理自体をインフラのJobQueueへ委譲することで、エラーハンドリングや進捗通知の重複を解消します。

## Goals / Non-Goals

**Goals:**
- `PersonaGenSlice` を「入力からLLMリクエスト配列（`[]llm_client.Request`）を生成する関数」と「LLMレスポンス配列（`[]llm_client.Response`）を受け取ってSQLiteにUPSERTする関数」に分割する。
- 内部で保持していたチャネル制御やLLMエラーハンドリング、再試行コードを破棄する。
- プロンプト生成時のトークン数計算や、作成済みペルソナの除外ロジック（`ExcludeAlreadyGenerated` 等のドメインフィルター）は完全に維持する。

**Non-Goals:**
- ペルソナ生成用プロンプトそのものの変更や、評価（Scorer）ロジックの変更。これらは既存ロジックを流用する。

## Decisions

1. **フェーズの分離**:
   - `PreparePrompts(ctx, input PersonaGenInput) ([]llm_client.Request, error)`: DBから作成済みを除外し、トークン計算を行い、純粋なプロンプトだけを返す。
   - `SaveResults(ctx, input PersonaGenInput, results []llm_client.Response) error`: パース成功した結果をSQLiteに保存する。

2. **通信と再試行の切り離し**:
   - 通信エラーによるリトライや、全体進捗のUI送信はインフラ層（JobQueueのワーカーおよび `ProgressNotifier`）に任せ、スライス内からは記述を完全に削除する。
