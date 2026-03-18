# Spec: e2e-required-scenarios/dictionary-builder

## Overview

`DictionaryBuilder` ページに対するページ単位 E2E の `必須シナリオ` を定義する。

## Requirements

### Requirement: DictionaryBuilder page SHALL define required scenarios for source browsing and search workflows
システムは `DictionaryBuilder` ページに対し、辞書ソース一覧の確認、選択後の詳細/編集導線、横断検索導線を `必須シナリオ` として定義しなければならない。これらのシナリオは、ページの主要な利用目的を代表する最小の統合シナリオでなければならない。

#### Scenario: DictionaryBuilder required scenarios are enumerated
- **WHEN** 開発者が `DictionaryBuilder` のページ単位 E2E 要件を確認する
- **THEN** 少なくとも一覧表示、選択後の詳細/編集導線、横断検索導線の 3 つが `必須シナリオ` として定義されていなければならない

### Requirement: DictionaryBuilder SHALL expose a required scenario for source list visibility
システムは `DictionaryBuilder` の必須シナリオとして、ページ初期表示時にインポート領域と登録済み辞書ソース一覧を確認できるシナリオを提供しなければならない。

#### Scenario: Source list is visible on page open
- **WHEN** Playwright E2E が `DictionaryBuilder` ページを開く
- **THEN** ページタイトル、XML インポート領域、登録済み辞書ソース一覧が表示されなければならない
- **AND** モックで供給した辞書ソースのファイル名または件数を識別できなければならない

### Requirement: DictionaryBuilder SHALL expose a required scenario for source selection to entries editor navigation
システムは `DictionaryBuilder` の必須シナリオとして、ソース選択から詳細表示を経てエントリ編集画面へ遷移できるシナリオを提供しなければならない。

#### Scenario: User selects source and opens entries editor
- **WHEN** Playwright E2E が辞書ソース一覧から対象ソースを選択する
- **THEN** `DetailPane` に対象ソースの詳細情報が表示されなければならない
- **AND** 利用者が `エントリを表示・編集` を操作すると `GridEditor` へ遷移し、対象ソースのエントリ一覧を識別できなければならない

### Requirement: DictionaryBuilder SHALL expose a required scenario for cross-search navigation
システムは `DictionaryBuilder` の必須シナリオとして、横断検索モーダルから検索結果画面へ遷移できるシナリオを提供しなければならない。

#### Scenario: User opens cross-search and sees result page
- **WHEN** Playwright E2E が `DictionaryBuilder` で横断検索モーダルを開き、検索キーワードを入力して実行する
- **THEN** 検索結果画面へ遷移し、検索クエリと結果件数を識別できなければならない
- **AND** 横断検索結果のエントリ一覧を確認できなければならない
