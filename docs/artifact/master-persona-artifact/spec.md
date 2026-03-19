# Master Persona アーティファクト

MasterPersona の final 成果物と task 中間生成物の保存境界を定義する。

## Requirements

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

#### Scenario: 詳細画面が final 成果物だけで完結する
- **WHEN** ユーザーが Persona 詳細を開く
- **THEN** システムは final 成果物から `persona_text`、`generation_request`、`dialogues`、メタ情報を返さなければならない
- **AND** 下書き保存物や queue の一時データを参照してはならない

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

### Requirement: MasterPersona artifact は lookup key DTO を公開しなければならない
`pkg/artifact/master_persona_artifact` は、`source_plugin` と `speaker_id` を使った lookup key DTO を自前で定義しなければならない。persona slice はこの lookup key DTO を受けて final 成果物 lookup を行い、workflow は slice 契約経由で利用しなければならない。

#### Scenario: persona slice が lookup key DTO で final 成果物を検索する
- **WHEN** persona slice が話者ペルソナを検索する
- **THEN** `master_persona_artifact` の lookup key DTO を使って final 成果物を検索しなければならない
- **AND** workflow が artifact repository を直接呼び出してはならない

### Requirement: MasterPersona の task 中間生成物は task 終了時に cleanup されなければならない
システムは、MasterPersona task が終了したら、`task_id` に紐づく中間生成物を cleanup しなければならない。cleanup は final 成果物を削除対象に含めてはならない。

#### Scenario: task 完了後に中間生成物を削除する
- **WHEN** final 成果物の保存後に task が完了する
- **THEN** システムは当該 `task_id` の中間生成物を削除しなければならない
- **AND** final 成果物は保持しなければならない

#### Scenario: task が失敗または中止で終了しても中間生成物を削除する
- **WHEN** task が failed または cancelled で終了する
- **THEN** システムは再利用不要となった当該 `task_id` の中間生成物を削除しなければならない
- **AND** 他 task の中間生成物や既存 final 成果物を削除してはならない
