## ADDED Requirements

### Requirement: MasterPersona の final 成果物は artifact の正本として保存されなければならない
システムは、生成済み MasterPersona の確定成果物を `pkg/artifact/master_persona_artifact` に保存しなければならない。画面表示と translation flow は同じ final 成果物を参照し、slice ローカル DB や task 中間生成物を正本として扱ってはならない。

#### Scenario: 生成済みペルソナを final 成果物として保存する
- **WHEN** MasterPersona の保存フェーズが成功する
- **THEN** システムは generated persona を `artifact` の final テーブルへ保存しなければならない
- **AND** 同一 `source_plugin + speaker_id` の既存 final がある場合は上書き設定に従って更新または保持しなければならない

#### Scenario: translation flow が final 成果物を再利用できる
- **WHEN** translation flow が話者ペルソナ lookup を要求する
- **THEN** slice は final 成果物を参照してペルソナを返さなければならない
- **AND** task 中間テーブルや queue 一時データを参照前提にしてはならない

### Requirement: MasterPersona の final 成果物は画面表示に必要な項目だけを保持しなければならない
MasterPersona の final 成果物は、現在の一覧画面と詳細画面が必要とする項目だけを保持しなければならない。対象は `persona_id`、`form_id`、`source_plugin`、`speaker_id`、`npc_name`、`editor_id`、`race`、`sex`、`voice_type`、`updated_at`、`persona_text`、`generation_request`、`dialogues` とする。`status`、`dialogue_count`、`dialogue_count_snapshot` は final 成果物に含めてはならない。

#### Scenario: final 成果物に status と dialogue count を持たない
- **WHEN** システムが final 成果物を保存または返却する
- **THEN** `status`、`dialogue_count`、`dialogue_count_snapshot` を保持してはならない
- **AND** 一覧や詳細 DTO にもそれらを必須項目として含めてはならない

### Requirement: MasterPersona の中間生成物は task 単位で分離保存されなければならない
システムは、下書き、生成リクエスト準備、再開制御のための中間生成物を final 成果物と分離し、`task_id` 単位で保存しなければならない。中間生成物は resume 期間中だけ保持され、UI の全件表示に使ってはならない。

#### Scenario: 中間生成物を task 単位で保存する
- **WHEN** `PreparePrompts` が下書きや request 準備用データを保存する
- **THEN** システムは `task_id` をキーに中間生成物を保存しなければならない
- **AND** final 成果物テーブルへ下書き状態の行を混在させてはならない

#### Scenario: 一覧取得は中間生成物を含めない
- **WHEN** ユーザーが MasterPersona 一覧を取得する
- **THEN** システムは final 成果物だけを返さなければならない
- **AND** 中間生成物を一覧へ混在させてはならない

### Requirement: persona slice は artifact lookup key DTO を利用しなければならない
`pkg/artifact/master_persona_artifact` は、`source_plugin` と `speaker_id` を使った lookup key DTO を自前で定義しなければならない。persona slice はこの lookup key DTO を受けて final 成果物 lookup を行い、workflow は slice 契約経由で利用しなければならない。

#### Scenario: persona slice が lookup key DTO で final 成果物を検索する
- **WHEN** persona slice が話者ペルソナを検索する
- **THEN** `master_persona_artifact` の lookup key DTO を使って final 成果物を検索しなければならない
- **AND** workflow が artifact repository を直接呼び出してはならない

## RENAMED Requirements

### Requirement: MasterPersona 一覧は生成済み成果物のみを表示しなければならない
FROM: `MasterPersona 一覧は保存済みダイアログ件数を表示しなければならない`
TO: `MasterPersona 一覧は生成済み成果物のみを表示しなければならない`

## MODIFIED Requirements

### Requirement: MasterPersona 一覧は生成済み成果物のみを表示しなければならない
MasterPersona 一覧は、下書きや task 中間生成物ではなく、final として確定した generated persona 成果物だけを表示しなければならない。システムは final 成果物から一覧 DTO を構築し、検索語とプラグイン絞り込みだけを適用しなければならない。一覧および詳細表示の DTO に `status`、`dialogueCount`、`dialogue_count_snapshot` を含めてはならない。

#### Scenario: 下書きは一覧に含まれない
- **WHEN** ユーザーが MasterPersona 一覧を開く
- **THEN** システムは generated の final 成果物だけを返さなければならない
- **AND** task 中間生成物や下書き保存物を一覧へ含めてはならない

#### Scenario: 一覧は検索語とプラグインで絞り込める
- **WHEN** ユーザーが一覧で検索語またはプラグイン絞り込みを指定する
- **THEN** システムは final 成果物の一覧に対してその条件だけを適用しなければならない
- **AND** status を前提にした絞り込み UI や DTO を要求してはならない

#### Scenario: 詳細表示は final 成果物だけで完結する
- **WHEN** ユーザーがペルソナ詳細を開く
- **THEN** システムは final 成果物から `persona_text`、`generation_request`、`dialogues`、メタ情報を返さなければならない
- **AND** 一覧表示のために `dialogueCount` を別計算したり、下書き取得を要求してはならない

## REMOVED Requirements

### Requirement: MasterPersona 一覧はステータスでフィルタできなければならない
