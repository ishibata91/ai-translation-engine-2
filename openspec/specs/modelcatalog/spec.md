# Model Catalog 仕様書

## Purpose
UI が LLM インフラへ直接依存せず、モデル一覧取得を単一スライス経由で行うための仕様を定義する。

## Requirements

### Requirement: UI は modelCatalog 経由でモデル一覧を取得しなければならない
フロントエンドは `llm` 実装へ直接アクセスせず、`modelCatalog` の公開 API を通してモデル一覧を取得しなければならない。`modelCatalog` は provider/endpoint/apiKey/namespace を入力として受け取り、内部で `config`・`secret`・`llmManager` を利用して取得を実行しなければならない。

#### Scenario: UI が provider 切替時に modelCatalog を呼び出す
- **WHEN** ユーザーがモデル設定画面で provider を切り替える
- **THEN** UI は `modelCatalog.ListModels(...)` を呼び出して候補を更新しなければならない
- **AND** UI は `llm` 契約や provider 固有実装を直接参照してはならない

### Requirement: モデル候補は ID 一意で返却されなければならない
`modelCatalog` は UI 選択値として利用可能な一意 ID を返却しなければならない。表示名重複がある場合でも ID をキーとして扱える形式で返却しなければならない。

#### Scenario: 同名モデルが複数ある場合でも UI が正しく描画できる
- **WHEN** provider から同一 display_name のモデルが複数返却される
- **THEN** `modelCatalog` は ID を基準に重複判定または識別可能な一覧を返す
- **AND** UI は option key/value に ID を使用して重複 key 警告を発生させてはならない
