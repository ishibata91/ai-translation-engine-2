## ADDED Requirements

### Requirement: ペルソナ保存は再開時も冪等でなければならない
MasterPersona の保存フェーズは、再試行または再開が発生しても同一 NPC に対して重複レコードを作成せず、upsert として確定保存しなければならない。

#### Scenario: 再開後の再保存で重複が作成されない
- **WHEN** 一部保存済み状態でタスクを再開する
- **THEN** システムは未保存分のみ新規反映し、既存 NPC レコードは更新として扱わなければならない

#### Scenario: 保存失敗 request だけ再試行される
- **WHEN** 保存フェーズで一部 request が失敗する
- **THEN** 次回再開では失敗 request のみ保存再試行し、成功済み request は再保存してはならない

### Requirement: リクエスト生成前の persona 下書き保存は原データ属性を正しく保持しなければならない
`PreparePrompts` 実行時に `npc_personas` / `npc_dialogues` へ事前保存する属性は、JSON抽出元の意味を変えずに保持しなければならない。`race` へ `record_type` 等の別属性を混入させてはならない。

#### Scenario: npc_personas の属性が抽出元と一致する
- **WHEN** リクエスト生成前に `npc_personas` へ下書き保存する
- **THEN** `race` には NPC の種族値のみが保存されなければならない
- **AND** `sex` / `voice_type` / `source_plugin` は空でなく保存されなければならない（入力に値がある場合）

#### Scenario: npc_dialogues の editor_id はレスポンス欠損時にフォールバックされる
- **WHEN** dialogue response の `editor_id` が空で、dialogue group 側に `editor_id` がある
- **THEN** `npc_dialogues.editor_id` には group 側 `editor_id` を保存しなければならない
