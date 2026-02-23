# TermTranslator Delta Spec

## MODIFIED Requirements

### Requirement: 翻訳プロンプトの生成フェーズ (Phase 1)
TermTranslatorSliceは、渡された入力データ（`TermTranslatorInput`）から辞書引きやコンテキスト構築を行い、インフラ層に渡すべきLLMリクエスト（`[]llm_client.Request`）の配列を返す純粋なドメイン関数として機能しなければならない。

#### Scenario: 翻訳リクエストの構築
- **WHEN** プロセスマネージャーから `TermTranslatorInput` 形式のテキストデータリストを受け取った
- **THEN** 内部の辞書検索（Tri-gram/FTS等）を行い、LLMプロンプトのコンテキストを構築する
- **AND** 構築されたプロンプトの配列 `[]llm_client.Request` を返す（自身ではLLMクライアントを呼び出さない）
- **AND** `specs/refactoring_strategy.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### Requirement: 翻訳結果の保存フェーズ (Phase 2)
TermTranslatorSliceは、コールバックとして渡されたLLMレスポンス（`[]llm_client.Response`）の配列をパースし、自身の永続化ストレージ（Mod用語DB等）に安全に保存しなければならない。

#### Scenario: 成功レスポンスのパースと保存
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから `TL: |にほんご|` フォーマットを抽出する
- **AND** パースに成功したテキストを Mod用語DB の該当レコードに対して UPSERT する
- **AND** `specs/refactoring_strategy.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: パース失敗・エラーレスポンスの処理
- **WHEN** レスポンスがエラー（`Success == false`）である、または `TL: |...|` 形式が含まれていない
- **THEN** 該当するレコードの更新を安全にスキップし、エラー詳細を構造化ログとして記録する
- **AND** 処理全体を中断せず、他の正常なレスポンスの処理を続行する

### Requirement: TraceID と 構造化ログの伝播
`specs/refactoring_strategy.md` の「構造化ログ基盤」セクションに従い、すべての操作は TraceID を追跡可能でなければならない。

#### Scenario: コンテキストの伝播
- **WHEN** 各フェーズの関数（`PreparePrompts`, `SaveResults`）が呼び出される
- **THEN** 第一引数として渡された `context.Context` を、DB操作や内部メソッド呼び出しに正しく伝播させる
- **AND** すべてのログ出力（`slog`）に `trace_id` および `span_id` が含まれることを保証する
