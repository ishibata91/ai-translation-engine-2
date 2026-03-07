## MODIFIED Requirements

### Requirement: ペルソナ保存は再開時も冪等でなければならない
MasterPersona の保存フェーズは、再試行または再開が発生しても同一 NPC を `source_plugin + speaker_id` で一意に識別し、重複レコードを作成せず UPSERT で確定保存しなければならない。`overwrite_existing` が有効な場合のみ既存レコードを更新し、無効な場合は既存を保持しなければならない。

#### Scenario: 再開後の再保存で重複が作成されない
- **WHEN** 一部保存済み状態でタスクを再開する
- **THEN** システムは未保存分のみ新規反映し、既存 NPC レコードは `source_plugin + speaker_id` 単位で更新または保持として扱わなければならない

#### Scenario: 保存失敗 request だけ再試行される
- **WHEN** 保存フェーズで一部 request が失敗する
- **THEN** 次回再開では失敗 request のみ保存再試行し、成功済み request は再保存してはならない

#### Scenario: 上書き有効時は既存行が更新される
- **WHEN** `overwrite_existing=true` で同一 `source_plugin + speaker_id` の保存対象が存在する
- **THEN** システムは既存行を更新し、重複行を新規作成してはならない

#### Scenario: 上書き無効時は既存行を保持する
- **WHEN** `overwrite_existing=false` で同一 `source_plugin + speaker_id` の保存対象が存在する
- **THEN** システムは既存行を保持し、既存行の内容を変更してはならない

### Requirement: リクエスト生成前の persona 下書き保存は原データ属性を正しく保持しなければならない
`PreparePrompts` 実行時に `npc_personas` / `npc_dialogues` へ事前保存する属性は、JSON抽出元の意味を変えずに保持しなければならない。`source_plugin` が欠損している場合は入力ファイル名から `*.esm|*.esl|*.esp` を抽出して補完し、抽出不能時は `UNKNOWN` を設定しなければならない。ペルソナ用途の `npc_dialogues` は原文中心の最小項目で保持し、訳文更新を前提としてはならない。

#### Scenario: npc_personas の属性が抽出元と一致する
- **WHEN** リクエスト生成前に `npc_personas` へ下書き保存する
- **THEN** `race` には NPC の種族値のみが保存されなければならない
- **AND** `sex` / `voice_type` / `source_plugin` は入力に値がある場合は空でなく保存されなければならない

#### Scenario: source_plugin 欠損時はファイル名から補完される
- **WHEN** 入力データの `source_plugin` が空で、入力パスまたは入力名に `hoge.esm` / `hoge.esl` / `hoge.esp` が含まれる
- **THEN** システムは該当拡張子付きファイル名を `source_plugin` として保存しなければならない

#### Scenario: source_plugin を補完できない場合は UNKNOWN を設定する
- **WHEN** `source_plugin` が空で、入力情報から `*.esm|*.esl|*.esp` を抽出できない
- **THEN** システムは `source_plugin` に `UNKNOWN` を保存しなければならない

#### Scenario: npc_dialogues は原文中心で保存される
- **WHEN** ペルソナ用途の会話データを `npc_dialogues` に保存する
- **THEN** システムは `source_text` を保存しなければならない
- **AND** `translated_text` の更新を必須処理として扱ってはならない
