## ADDED Requirements

### Requirement: MasterPersona は task 終了時に中間生成物 cleanup を完了しなければならない
MasterPersona workflow は、task が終了したら `task_id` に紐づく persona 中間生成物 cleanup を persona slice 経由で完了しなければならない。workflow は artifact repository を直接参照してはならず、cleanup は必ず slice 契約を通じて実行しなければならない。

#### Scenario: 保存成功後に cleanup を実行する
- **WHEN** MasterPersona workflow が final persona の保存を完了して task を完了させる
- **THEN** システムは task 完了前に persona slice の cleanup 契約を呼び出さなければならない
- **AND** 当該 `task_id` の中間生成物を削除しなければならない

#### Scenario: 失敗または中止で終了する task でも cleanup を実行する
- **WHEN** MasterPersona task が failed または cancelled で終了する
- **THEN** システムは persona slice 経由で当該 `task_id` の中間生成物 cleanup を実行しなければならない
- **AND** workflow から artifact repository を直接呼び出してはならない

#### Scenario: cleanup は final 成果物を削除しない
- **WHEN** workflow が cleanup を実行する
- **THEN** システムは task スコープの中間生成物だけを削除しなければならない
- **AND** final 成果物や他 task の中間生成物を削除してはならない
