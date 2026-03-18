# Purpose

翻訳フローのデータロード工程における phase 順序と artifact handoff を定義する。

## Requirements

### Requirement: Translation Flow workflow はデータロードを先頭 phase として扱わなければならない
システムは、`TranslationFlow` の workflow において `データロード` を先頭 phase として扱い、`用語`、`ペルソナ`、`要約`、`翻訳`、`エクスポート` より前に配置しなければならない。

#### Scenario: 翻訳フローの開始順序を決定する
- **WHEN** workflow が翻訳フローの phase 順序を決定する
- **THEN** `データロード` は最初の phase として扱われなければならない
- **AND** 有効なロード結果を持たないまま `用語` 以降へ進行してはならない

### Requirement: Translation Flow workflow はロード済みデータを artifact 境界で受け渡さなければならない
システムは、データロードで得られたパース済みデータを UI の一時 state や slice ローカル保存へ閉じ込めず、`artifact` の識別子と検索条件を用いて後続 phase へ受け渡さなければならない。

#### Scenario: ロード結果を後続 phase へ引き渡す
- **WHEN** workflow がデータロード完了後に `用語` phase 以降を開始する
- **THEN** workflow は `task_id`、file 識別子、section 識別子など artifact 境界の情報を束ねて後続 phase を呼び出さなければならない
- **AND** UI ローカル state や parser の内部 DTO をそのまま後続 phase へ渡してはならない

#### Scenario: 既存 task を再表示する
- **WHEN** ユーザーが既存の翻訳 task を開き直す
- **THEN** workflow は artifact に保存されたロード結果を用いてデータロード phase を復元できなければならない
- **AND** 同じファイル群を選び直さないと先へ進めない設計にしてはならない
