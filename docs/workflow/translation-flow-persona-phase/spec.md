# 翻訳フロー ペルソナ生成

翻訳フローの `ペルソナ生成` phase における候補投影、既存 Master Persona 再利用、実行制御を定義する。

## Requirements

### Requirement: Translation Flow workflow は persona phase を translation project task 配下で実行しなければならない
システムは、`TranslationFlow` の persona phase を `translation_project` task 配下の phase として実行しなければならない。候補一覧、実行、resume、retry は同じ `task_id` を使って追跡され、`単語翻訳` の後、`要約` の前に位置付けられなければならない。

#### Scenario: persona phase が同一 task ID で追跡される
- **WHEN** workflow が translation flow の persona phase を開始または再表示する
- **THEN** システムは同じ `translation_project.task_id` を使って preview、実行、resume を追跡しなければならない
- **AND** 別 task を新規発行してはならない

### Requirement: workflow は `source_plugin + speaker_id` で既存 Master Persona を除外した候補集合を使わなければならない
システムは、translation input artifact から抽出した NPC 候補を `source_plugin + speaker_id` で正規化し、`master_persona_artifact` final に既存 persona がある候補を `既存 Master Persona` として除外しなければならない。preview と execute は同じ候補 planner を使い、除外結果がずれてはならない。

#### Scenario: preview と execute が同じ候補集合を使う
- **WHEN** ユーザーが persona target 一覧を開いた後に phase を開始する
- **THEN** workflow は同じ候補 planner を使って preview 行と request 対象を決定しなければならない
- **AND** preview で `既存 Master Persona` だった候補を execute で request 化してはならない

#### Scenario: 重複候補は 1 件に統合される
- **WHEN** 同じ `source_plugin + speaker_id` を持つ NPC が複数ファイルまたは複数行に存在する
- **THEN** workflow はそれらを 1 候補へ統合しなければならない
- **AND** 重複 request を生成してはならない

### Requirement: workflow は `新規生成数 0 件` のとき no-op 完了しなければならない
システムは、候補集合のすべてが既存 Master Persona で解決できる場合、runtime を呼ばずに persona phase を完了しなければならない。summary は `新規生成 0 件` と `再利用済み` を返し、失敗扱いにしてはならない。

#### Scenario: 全件再利用で runtime を呼ばない
- **WHEN** persona phase の `pending_generation` が 0 件である
- **THEN** workflow は LLM 実行を開始してはならない
- **AND** phase を no-op 完了として summary に反映しなければならない

### Requirement: workflow は partial failure 後に未解決候補だけを retry しなければならない
システムは、persona phase の一部 request が失敗した場合、成功済み行と既存 Master Persona を保持したまま、未解決候補だけを retry しなければならない。resume 時も final 成果物と task summary を基に同じ state を復元しなければならない。

#### Scenario: partial failed で failed 行だけを再送する
- **WHEN** いくつかの persona request が失敗した後に再試行する
- **THEN** workflow は failed または未生成の候補だけを再度 request 化しなければならない
- **AND** 既に final 成果物へ保存済みの候補を再送してはならない

#### Scenario: 再表示時に row state を復元する
- **WHEN** ユーザーが既存 translation task を開き直す
- **THEN** workflow は final 成果物と task summary を使って `既存 Master Persona` `生成済み` `生成失敗` を復元しなければならない
- **AND** stale な UI state だけを根拠に復元してはならない

### Requirement: persona phase の NPC 一覧は MasterPersona 一覧と同じ table shell を使わなければならない
システムは、translation flow の persona phase に表示する NPC 一覧を、MasterPersona 一覧と同じ table shell、列順、選択ハイライトで表示しなければならない。`speaker_id` は一覧上 `FormID` 列として扱い、`source_plugin` と `npc_name` を同じ主列で表示しなければならない。一方で phase 固有の `既存 Master Persona` `生成対象` `生成中` `生成済み` `生成失敗` は補助表示として識別できなければならない。

#### Scenario: persona phase 一覧が MasterPersona と同じ表形式で描画される
- **WHEN** ユーザーが translation flow の persona phase を開く
- **THEN** システムは MasterPersona 一覧と同じ table shell と選択スタイルで NPC 行を描画しなければならない
- **AND** `speaker_id` を `FormID` 列へ表示しなければならない

#### Scenario: row state を保ったまま shared list contract で表示できる
- **WHEN** persona target row が `reused` `pending` `running` `generated` `failed` のいずれかで返る
- **THEN** システムは shared list contract の主列を崩さずに phase state を補助表示で識別できなければならない
- **AND** row state 可視性を詳細ペインだけへ押し込めてはならない
