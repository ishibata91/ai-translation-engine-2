# フォーマット境界

## Purpose

外部フォーマット（ゲーム固有/ツール固有）の解釈・生成責務を `pkg/format` に集約し、workflow や他責務区分から分離するための共通仕様を定義する。

## Requirements

### Requirement: 外部フォーマット適応は pkg/format 配下へ配置されなければならない
システムは、ゲーム固有またはツール固有の入出力フォーマットを解釈・生成する実装を `pkg/format` 配下へ配置しなければならない。workflow、slice、gateway は外部フォーマットへの適応責務を自区分へ抱え込んではならない。

#### Scenario: Skyrim parser と xTranslator exporter を配置する
- **WHEN** 開発者が Skyrim 抽出 JSON の入力処理と xTranslator XML の出力処理を実装する
- **THEN** Skyrim parser は `pkg/format/parser/skyrim` 配下に配置されなければならない
- **AND** xTranslator exporter は `pkg/format/exporter/xtranslator` 配下に配置されなければならない

### Requirement: format 実装は既存 workflow 契約へ接続されなければならない
システムは、`pkg/format` 配下へ移設した実装を workflow から既存の `Parser` / `Exporter` 契約名で利用できなければならない。format 境界化を理由に workflow 公開契約名を変更してはならない。

#### Scenario: workflow が format 実装を利用する
- **WHEN** workflow が抽出データのロードまたは XML エクスポートを呼び出す
- **THEN** workflow は既存の `Parser` / `Exporter` 契約名のまま format 実装を解決できなければならない
- **AND** 具象実装の知識は composition root または provider に閉じ込められなければならない

### Requirement: format 移設時は旧配置前提の補助コマンドを残してはならない
システムは、`pkg/format` への移設完了後に旧配置前提の補助コマンドや互換 import を残してはならない。不要になった補助コマンドは削除されなければならない。

#### Scenario: 旧 parser 補助コマンドを見直す
- **WHEN** `pkg/slice/parser` から `pkg/format/parser/skyrim` へ実装を移設する
- **THEN** 旧配置前提の `cmd/parser` は削除されなければならない
- **AND** 一時的な互換 import を恒久運用として残してはならない
