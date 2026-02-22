# TermTranslator Delta Spec

## MODIFIED Requirements

### Requirement: 翻訳プロンプトの生成フェーズ (Phase 1)
TermTranslatorSliceは、渡された入力データ（`TermTranslatorInput`）から辞書引きやコンテキスト構築を行い、インフラ層に渡すべきLLMリクエスト（`[]llm_client.Request`）の配列を返す纯関数として機能しなければならない。

#### Scenario: 翻訳リクエストの構築
- **WHEN** プロセスマネージャーから `TermTranslatorInput` 形式のテキストデータリストを受け取った
- **THEN** 内部の辞書検索（Tri-gram/FTS等）を行い、LLMプロンプトのコンテキストを構築する
- **AND** 構築されたプロンプトの配列 `[]llm_client.Request` を返す（自身ではLLMクライアントを呼び出さない）

### Requirement: 翻訳結果の保存フェーズ (Phase 2)
TermTranslatorSliceは、コールバックとして渡されたLLMレスポンス（`[]llm_client.Response`）の配列をパースし、自身の永続化ストレージ（Mod用語DB等）に安全に保存しなければならない。

#### Scenario: 成功レスポンスのパースと保存
- **WHEN** プロセスマネージャーから、自身の生成したリクエストに対応する `[]llm_client.Response` が渡された
- **THEN** 各レスポンスから `TL: |にほんご|` フォーマットを抽出する
- **AND** パースに成功したテキストを Mod用語DB の該当レコードに対して UPSERT する
- **AND** （失敗レスポンスが含まれていた場合はそれを安全にスキップ・ログ出力する）
