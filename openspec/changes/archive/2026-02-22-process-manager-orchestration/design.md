# Design: Process Manager Orchestration

## Context
VSA (Vertical Slice Architecture) において、各スライスは独立したコンテキストを保ちますが、それらを組み合わせてエンドツーエンドの「翻訳プロセス」を構成するためには、上位の調整役が必要です。
`ProcessManager` は、ユーザーのリクエストを受け取り、各スライスの DTO にデータをマッピングし、JobQueue への委譲と結果の回収をオーケストレートする「スライス間のハブ」として機能します。

## Goals / Non-Goals

**Goals:**
- **Anti-Corruption Layer (ACL)**: `LoaderSlice` の出力 DTO から後続スライス（`TermTranslator` 等）の入力 DTO へのマッピングを担う。
- **Process Orchestration**: スライスの「Phase 1 (プロンプト生成)」と「Phase 2 (結果保存)」を順次呼び出す。
- **Job Delegation**: LLM実行を `JobQueue` に委譲し、完了結果をスライスにコールバックする。
- **UI Interaction**: WebSocket/SSE ハンドラを提供し、進行状態（State）と進捗（Progress）をリアルタイムに UI へ送信する。
- **Resumability**: 再起動時、実行中の JobQueue（特に Batch API）の状態を復旧し、処理を再開できる。

**Non-Goals:**
- 特定のドメインロジック（翻訳アルゴリズム等）の保持。
- JobQueue 内部の再試行・エラーハンドリングの詳細実装。

## Decisions

1. **SQLite による進行状態 (State) の永続化**:
   - `process_state.db` を使用し、現在実行中の `ProcessID`, `TargetSlice`, `InputFile` (入力JSON名), `BatchJobID` (Batch API 使用時), `CurrentPhase` を記録します。
   - これにより、アプリケーションが再起動しても、「どのファイルを、どのスライスで、どこまで処理したか」を完全に復元でき、JobQueue インフラで動いている長いバッチジョブを紛失せず、結果回収から再開できます。

2. **同一ファイル内での SRP 分割 (Method-Level SRP)**:
   - `refactoring_strategy.md` セクション 5 に従い、オーケストレーターの肥大化対策として「ファイル分割」ではなく「プライベートメソッドによる分割」を徹底します。
   - `ExecuteTermTranslation`, `handleBatchCompletion`, `mapToTranslatorInput` 等、意味のある単位で分割し、メイン関数は目次のような構造にします。

3. **共有モデルの排除とマッピング**:
   - `pkg/domain` レベルの共有 DTO への依存を排除します。
   - すべてのデータ受け渡しは `ProcessManager` が担い、コンシューマ（各スライス）が要求する `InputDTO` へ型変換を行います。

## Risks / Trade-offs

- **[Risk] オーケストレーターの結合度**: 多くのスライスを知る必要がある。
  - **Mitigation**: インターフェース（Contract）にのみ依存し、各スライスの内部実装（Implementation）には一切関知しないことで、変更の影響範囲を局所化します。
