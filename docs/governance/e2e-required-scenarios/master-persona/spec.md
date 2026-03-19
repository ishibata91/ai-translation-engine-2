# Master Persona E2E 必須シナリオ

## Overview

`MasterPersona` ページに対するページ単位 E2E の `必須シナリオ` を定義する。

## Requirements

### Requirement: MasterPersona page SHALL define required scenarios for persona browsing, prompt editing, model configuration, and task start workflows
システムは `MasterPersona` ページに対し、初期表示、NPC 一覧から詳細確認、プロンプト設定の表示 / 編集、モデル設定の主要操作、JSON 選択後のタスク開始導線を `必須シナリオ` として定義しなければならない。これらのシナリオは、ページの主要な利用目的を代表する最小の統合シナリオでなければならない。

#### Scenario: MasterPersona required scenarios are enumerated
- **WHEN** 開発者が `MasterPersona` のページ単位 E2E 要件を確認する
- **THEN** 少なくとも初期表示、NPC 詳細確認、プロンプト設定、モデル設定、タスク開始導線の 5 つが `必須シナリオ` として定義されていなければならない

### Requirement: MasterPersona SHALL expose a required scenario for page initialization visibility
システムは `MasterPersona` の必須シナリオとして、ページ初期表示時にタイトル、JSON 選択領域、進捗表示、プロンプト設定カード、モデル設定領域を確認できるシナリオを提供しなければならない。

#### Scenario: MasterPersona page shows initial controls on open
- **WHEN** Playwright E2E が `MasterPersona` ページを開く
- **THEN** ページタイトル、JSON 選択ボタン、全体進捗表示、ユーザープロンプト / システムプロンプトのカード、モデル設定セクションが表示されなければならない

### Requirement: MasterPersona SHALL expose a required scenario for NPC list to detail inspection
システムは `MasterPersona` の必須シナリオとして、NPC 一覧から対象 NPC を選択し、右ペインでペルソナ詳細を確認できるシナリオを提供しなければならない。

#### Scenario: User selects NPC and sees persona detail
- **WHEN** Playwright E2E が `MasterPersona` ページでモックされた NPC 一覧から対象 NPC を選択する
- **THEN** 右側の詳細ペインに対象 NPC 名と FormID が表示されなければならない
- **AND** ペルソナ情報または関連タブの内容を識別できなければならない

### Requirement: MasterPersona SHALL expose a required scenario for prompt setting card visibility and editability boundaries
システムは `MasterPersona` の必須シナリオとして、`PromptSettingCard` によりユーザープロンプトが編集可能であり、システムプロンプトが read only であることを確認できるシナリオを提供しなければならない。

#### Scenario: User identifies editable and read-only prompt cards
- **WHEN** Playwright E2E が `MasterPersona` ページのプロンプト設定領域を確認する
- **THEN** `ユーザープロンプト` カードは編集可能として識別できなければならない
- **AND** `システムプロンプト` カードは read only として識別できなければならない

#### Scenario: User edits prompt text in editable card
- **WHEN** Playwright E2E が `ユーザープロンプト` カードの内容を変更する
- **THEN** 変更したテキストがテキストエリアに反映されなければならない
- **AND** `システムプロンプト` カードの内容や read only 状態に影響を与えてはならない

### Requirement: MasterPersona SHALL expose a required scenario for model settings interaction
システムは `MasterPersona` の必須シナリオとして、`ModelSettings` によりプロバイダ切替、モデル選択候補表示、主要スライダー値の変化を確認できるシナリオを提供しなければならない。

#### Scenario: User changes provider and sees model settings update
- **WHEN** Playwright E2E が `ModelSettings` で AI プロバイダを切り替える
- **THEN** 選択中プロバイダの値が更新されなければならない
- **AND** モデル候補または関連入力欄が切り替え後の状態として識別できなければならない

#### Scenario: User adjusts primary model controls
- **WHEN** Playwright E2E が `ModelSettings` の主要スライダーまたは入力欄を操作する
- **THEN** 並列実行数、Temperature、またはコンテキスト長の表示値のうち対象コントロールに対応する値が更新されなければならない

### Requirement: MasterPersona SHALL expose a required scenario for JSON selection to task start readiness
システムは `MasterPersona` の必須シナリオとして、JSON ファイル選択後に開始ボタンが有効化され、タスク開始操作の反映を確認できるシナリオを提供しなければならない。

#### Scenario: User selects JSON and starts persona generation task
- **WHEN** Playwright E2E が `MasterPersona` ページで JSON ファイルを選択し、`新規タスク開始` を操作する
- **THEN** 選択した JSON パスが表示されなければならない
- **AND** 開始ボタン押下後に `生成中...`、進行中表示、または開始済み状態のいずれかを識別できなければならない
