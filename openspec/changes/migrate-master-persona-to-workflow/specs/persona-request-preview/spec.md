## MODIFIED Requirements

### Requirement: MasterPersona 画面の開始ボタンは task 境界を起動しなければならない
`MasterPersona` 画面の開始ボタンは、controller を通じて workflow 管理下の MasterPersona タスク開始要求を送らなければならない。開始時入力には `source_json_path` と `overwrite_existing` を含めなければならない。

#### Scenario: UI が workflow 管理下の開始要求を送る
- **WHEN** ユーザーが MasterPersona 画面で開始ボタンを押す
- **THEN** システムは controller を通じて workflow 管理下の開始要求を送らなければならない
- **AND** 同一 task ID で進捗・完了・失敗を追跡できなければならない
