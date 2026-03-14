## MODIFIED Requirements

### Requirement: 独立性: エクスポート用データの受け取りと独自DTO定義
xTranslator XML 出力機能は、`pkg/format/exporter/xtranslator` 配下の format adapter として実装されなければならない。システムは、他スライスの DTO に直接依存せず、workflow から受け取った本機能専用 DTO を用いて XML ファイルの構築と保存を完結しなければならない。

#### Scenario: 独自定義DTOによる初期化とエクスポート処理
- **WHEN** workflow から本機能専用の入力 DTO（`ExportInput`）が提供された場合
- **THEN** `pkg/format/exporter/xtranslator` の実装は外部パッケージの DTO に依存せず、XML ファイルの構築と保存を完結できなければならない
- **AND** workflow から見える契約名は `Exporter` のまま維持されなければならない

### Requirement: メインインターフェース
システムは、xTranslator XML 出力機能を workflow から `Exporter` 契約として利用できなければならない。provider または composition root は `pkg/format/exporter/xtranslator` の具象実装を `Exporter` 契約へ束縛しなければならない。

#### Scenario: workflow が xTranslator exporter を解決する
- **WHEN** アプリケーションが XML エクスポート実行のために exporter を初期化する
- **THEN** provider または composition root は `pkg/format/exporter/xtranslator` の具象型を `Exporter` 契約として解決しなければならない
- **AND** workflow は format 配下の具象型名を直接参照してはならない
