## ADDED Requirements

### Requirement: ペルソナ保存は再開時も冪等でなければならない
MasterPersona の保存フェーズは、再試行または再開が発生しても同一 NPC に対して重複レコードを作成せず、upsert として確定保存しなければならない。

#### Scenario: 再開後の再保存で重複が作成されない
- **WHEN** 一部保存済み状態でタスクを再開する
- **THEN** システムは未保存分のみ新規反映し、既存 NPC レコードは更新として扱わなければならない

#### Scenario: 保存失敗 request だけ再試行される
- **WHEN** 保存フェーズで一部 request が失敗する
- **THEN** 次回再開では失敗 request のみ保存再試行し、成功済み request は再保存してはならない
