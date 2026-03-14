## MODIFIED Requirements

### Requirement: Parser スライス DI プロバイダー
Skyrim 抽出 JSON を扱う parser 機能は、`pkg/format/parser/skyrim` 配下の format adapter として実装されなければならない。システムは、Google Wire または同等の composition root を介して依存性注入プロバイダー関数を提供し、workflow からは `Parser` 契約として解決できなければならない。

#### Scenario: DI 初期化
- **WHEN** アプリケーションが抽出データ読み込み用の parser を初期化する場合
- **THEN** 内部構造体を直接インスタンス化することなく、`pkg/format/parser/skyrim` の provider を通じて `Parser` 契約を解決しなければならない
- **AND** workflow は interface 名を変更せずに新配置の実装を利用できなければならない

### Requirement: 内部プロセスのカプセル化
Skyrim parser における順次ロードと並列デコードの手順は、`pkg/format/parser/skyrim` の `Parser` 実装内にカプセル化され、workflow に公開されてはならない。workflow は最終的な `ExtractedData` またはエラーのみを受け取らなければならない。

#### Scenario: ロードプロセスのオーケストレーション
- **WHEN** workflow が `LoadExtractedJSON` を呼び出す場合
- **THEN** `pkg/format/parser/skyrim` の実装はファイルの読み込み、デコード、構造化を内部で調整し、最終的な `ExtractedData` またはエラーのみを返さなければならない
- **AND** workflow は format 配下の内部処理手順へ依存してはならない
