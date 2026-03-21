# 翻訳フロー ペルソナ生成 UI

翻訳フローの `ペルソナ生成` タブにおける表示責務、状態遷移、ユーザー操作を定義する。

## Requirements

### Requirement: Translation Flow UI は `単語翻訳` の次に `ペルソナ生成` phase を表示しなければならない
システムは、`TranslationFlow` の UI において `ペルソナ生成` を `単語翻訳` の次、`要約` の前に表示しなければならない。ユーザーは同じ translation project task の流れの中で persona phase を開き、既存 task 再表示時も同じ phase として復元できなければならない。

#### Scenario: タブ順序に persona phase が含まれる
- **WHEN** ユーザーが翻訳フロー画面を開く
- **THEN** UI は `データロード` `単語翻訳` `ペルソナ生成` `要約` の順で phase を表示しなければならない
- **AND** `ペルソナ生成` は独立したタブ内容を持たなければならない

### Requirement: ペルソナ生成タブは既存 Master Persona 再利用数と新規生成数を分けて表示しなければならない
システムは、persona phase の要約カードと NPC 一覧において、`検出 NPC 数` `既存 Master Persona 再利用数` `新規生成数` を分けて表示しなければならない。`既存 Master Persona` 行は一覧へ表示しても `新規生成数` に含めてはならない。

#### Scenario: 既存 Master Persona 行が一覧に残る
- **WHEN** persona phase に `既存 Master Persona` と `新規生成対象` が混在している
- **THEN** UI は両方の行を同じ一覧で表示しなければならない
- **AND** `既存 Master Persona` 行には再利用 badge を表示しなければならない
- **AND** `新規生成数` には再利用行を含めてはならない

#### Scenario: 全件再利用では no-op 完了を表示する
- **WHEN** 検出された NPC 全件に既存 final persona が存在する
- **THEN** UI は `新規生成 0 件` と `既存 Master Persona を再利用します` を表示しなければならない
- **AND** `ペルソナ生成を開始` を無効にしなければならない
- **AND** `次へ` を即時有効にしなければならない

### Requirement: ペルソナ生成タブは選択 NPC の状態に応じた詳細を表示しなければならない
システムは、一覧で選択された NPC に応じて、cached 行では既存 persona 本文を、pending 行では会話抜粋とメタ情報を、generated 行では保存済み persona 本文を表示しなければならない。UI は persona の有無を状態に応じて観測可能にしなければならない。

#### Scenario: cached 行を選択すると既存 persona が見える
- **WHEN** ユーザーが `既存 Master Persona` 行を選択する
- **THEN** UI は final 成果物から取得した persona 本文を read-only で表示しなければならない
- **AND** 当該行を `未生成` と見せてはならない

#### Scenario: pending 行を選択すると未生成理由を示す
- **WHEN** ユーザーが `生成対象` 行を選択する
- **THEN** UI は会話抜粋、source plugin、speaker/editor 識別情報を表示しなければならない
- **AND** persona 本文の代わりに `この NPC はまだ生成されていません` を示さなければならない

### Requirement: ペルソナ生成タブは新規生成対象だけを開始し、失敗時は retry と続行可否を明示しなければならない
システムは、persona phase 実行中に `既存 Master Persona` 行を request 対象に含めず、新規生成対象だけを実行しなければならない。実行結果が一部失敗した場合、UI は `生成失敗` 行を識別可能にしたうえで `再試行` を提示し、同時に後続 phase へ進むと persona 欠損行が残ることを明示しなければならない。

#### Scenario: 実行中は pending 行だけが進捗対象になる
- **WHEN** ユーザーが `ペルソナ生成を開始` を押す
- **THEN** UI は `生成対象` 行だけを `生成中` または `生成済み` へ遷移させなければならない
- **AND** `既存 Master Persona` 行は再利用状態のまま保持しなければならない

#### Scenario: partial failed では retry と next を同時に提示する
- **WHEN** 新規生成対象の一部だけが失敗して phase が `partialFailed` になる
- **THEN** UI は `生成失敗` 行を一覧で識別できなければならない
- **AND** `再試行` を有効にしなければならない
- **AND** `次へ` は有効にしつつ、失敗行が persona なしで後続 phase に進むことを明示しなければならない
