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

1. **フェーズの分離 (Vertical Slice Architecture 準拠)**:
   - `PreparePrompts(ctx context.Context, input PersonaGenInput) ([]llm_client.Request, error)`: DBから作成済みを除外し、トークン計算を行い、純粋なプロンプトだけを返す。
   - `SaveResults(ctx context.Context, input PersonaGenInput, results []llm_client.Response) error`: パース成功した結果をSQLiteに保存する。
   - **DTOの完全分離**: 他のスライスとDTOを共有せず、`PersonaGenInput` はこのスライス固有の定義（`contract.go` 等）として自己完結させる。

2. **通信と再試行の切り離し**:
   - 通信エラーによるリトライや、全体進捗のUI送信はインフラ層（JobQueueのワーカーおよび `ProgressNotifier`）に任せ、スライス内からは記述を完全に削除する。

3. **構造化ログ (slog + OpenTelemetry) の義務付け**:
   - 全ての公開メソッド（`PreparePrompts`, `SaveResults`）の入り口と出口で、`slog.DebugContext` を用いて引数と戻り値を記録する。
   - `context.Context` を伝播させ、`trace_id` が自動付与されるようにする。

4. **テスト戦略 (VSA準拠)**:
   - スライス単位の網羅的なパラメタライズドテスト（Table-Driven Test）を実装する。
   - DB操作を含むため、テスト専用の SQLite (:memory:) を使用する。
   - 細粒度なユニットテストは AI の自由度を奪うため排除し、スライスレベルの仕様検証に集中する。

5. **LLM出力フォーマットの統一**:
   - プロンプト要件として、出力フォーマットを `TL: |ペルソナ内容|` 形式に指定する。
   - `SaveResults` 内でのパース時は、パイプ区切り形式から安定して結果を抽出するフォールバックロジックを実装する。

## 6. SaveResults 詳細設計

`SaveResults` は、JobQueueによって実行されたLLMの応答（`[]llm_client.Response`）をドメイン知識に基づいて解釈し、最終的な永続化を行う責務を持ちます。

### 6.1. レスポンスの紐付け
- `llm_client.Request` 生成時に `Metadata` フィールドへ対象NPCの識別子（`InternalID` 等）を格納します。
- `SaveResults` では `Response.Metadata` からこの識別子を取り出し、どのNPCのペルソナであるかを確定させます。

### 6.2. ペルソナ抽出アルゴリズム (Parsing)
LLMの出力は以下の優先順位でパースします。
1. **正規表現による抽出**: `TL:\s*\|(.*?)\|` を検索し、最初のグループを取得します。
2. **フォールバック1 (プレフィックス除去)**: `TL:` で始まる行を探し、その後の文字列をトリミングします。
3. **フォールバック2 (全体トリミング)**: パイプ記号 `|` が含まれる場合、その間のテキストを抽出します。
抽出されたテキストが空、または著しく短い（例: 5文字以下）場合は、パース失敗としてログを記録し、そのレコードの保存をスキップします。

### 6.3. データベース保存 (UPSERT Logic)
- `ModTermStore` インターフェースを介して保存を実行します。
- 保存先テーブル: `mod_terms`
- カラムマッピング:
    - `term_key`: `NPC_PERSONA_<NPC_ID>` (命名規則に従う)
    - `original_text`: 空（またはソースとなる要約の一部）
    - `translated_text`: 抽出されたペルソナ文
    - `category`: `Persona`
- すでに同一 `term_key` が存在する場合は、新しいペルソナで上書き（UPSERT）します。

### 6.4. ログとエラーハンドリング
- 保存に成功したNPCのIDと、パースに失敗したNPCのIDをそれぞれ `slog.InfoContext` / `slog.WarnContext` で出力します。
- 1件のDBエラーで全体の保存をロールバックするのではなく、エラーが発生したレコードのみをスキップし、可能な限り多くのデータを保存します。
