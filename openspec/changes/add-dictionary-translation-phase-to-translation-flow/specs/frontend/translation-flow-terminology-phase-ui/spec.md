## ADDED Requirements

### Requirement: Translation Flow UI は単語翻訳を独立した phase として表示しなければならない
システムは、translation flow UI において `単語翻訳` を `データロード` の次に並ぶ独立 phase として表示しなければならない。step 表示とメインコンテンツは現在の phase と一致し、未実行時・実行中・完了時の状態を識別できなければならない。

#### Scenario: データロード後に単語翻訳 phase を表示する
- **WHEN** ユーザーがデータロードを完了した translation flow task を開く
- **THEN** UI は次の phase として `単語翻訳` を表示しなければならない
- **AND** `本文翻訳` を `単語翻訳` より先に操作可能にしてはならない

### Requirement: Translation Flow UI は単語翻訳の実行サマリを表示しなければならない
システムは、単語翻訳 phase 完了後に `対象件数 / 保存件数 / 失敗件数` を表示しなければならない。

#### Scenario: 単語翻訳結果を確認する
- **WHEN** ユーザーが単語翻訳 phase の結果画面を閲覧する
- **THEN** UI は対象件数と保存件数を表示しなければならない
- **AND** 保存失敗がある場合は失敗件数を識別できなければならない

### Requirement: Translation Flow UI は単語翻訳の進行状態を表示しなければならない
システムは、単語翻訳 phase の実行中に terminology ジョブ生成中、翻訳実行中、保存中、完了の状態を表示できなければならない。

#### Scenario: 単語翻訳の進行を確認する
- **WHEN** ユーザーが単語翻訳 phase を実行する
- **THEN** UI は現在の進行状態を表示しなければならない
- **AND** 実行中であることを本文翻訳 phase と区別できなければならない

### Requirement: Translation Flow UI は terminology 用の LLM リクエスト設定を編集できなければならない
システムは、translation flow の単語翻訳 phase において、operator が provider、model、execution profile、temperature、endpoint、apiKey、contextLength、syncConcurrency などの LLM リクエスト設定を `ModelSettings` ベースの UI で編集できなければならない。設定は master persona と別 namespace で保持しなければならない。

#### Scenario: 単語翻訳用の request 設定を編集する
- **WHEN** ユーザーが単語翻訳 phase の設定パネルを開く
- **THEN** UI は terminology 用 request 設定を表示し、編集内容を保存できなければならない
- **AND** master persona の request 設定を上書きしてはならない

### Requirement: Translation Flow UI は terminology 用 prompt を編集できなければならない
システムは、translation flow の単語翻訳 phase において、operator が system prompt と user prompt を `PromptSettingCard` ベースの UI で編集できなければならない。prompt 編集内容は再訪時に復元されなければならない。

#### Scenario: 単語翻訳用 prompt を調整する
- **WHEN** ユーザーが terminology phase の prompt 設定を編集する
- **THEN** UI は system prompt と user prompt の両方を編集できなければならない
- **AND** 編集内容は次回表示時にも復元されなければならない

### Requirement: Translation Flow UI は再訪時に単語翻訳結果と設定を復元表示しなければならない
システムは、既存 task を再表示した場合でも保存済み単語翻訳結果を用いて phase 状態とサマリを復元表示しなければならない。あわせて terminology 用 request 設定と prompt 設定も前回保存値から復元されなければならない。

#### Scenario: 途中で画面を離れて戻る
- **WHEN** ユーザーが単語翻訳完了後に別画面へ移動して再度 translation flow に戻る
- **THEN** UI は最後に保存された単語翻訳結果と設定値を再表示しなければならない
- **AND** ユーザーに再実行だけを要求してはならない
