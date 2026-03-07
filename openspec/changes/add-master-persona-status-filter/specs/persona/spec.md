## MODIFIED Requirements

### Requirement: リクエスト生成前の persona 下書き保存は原データ属性を正しく保持しなければならない
`PreparePrompts` 実行時に `npc_personas` / `npc_dialogues` へ事前保存する属性は、JSON抽出元の意味を変えずに保持しなければならない。`race` へ `record_type` 等の別属性を混入させてはならない。`source_plugin` が欠損している場合は入力ファイル名から `*.esm|*.esl|*.esp` を補完し、抽出不能時は `UNKNOWN` を設定しなければならない。加えて、`npc_personas.status` には英語値 `draft` を保存し、リクエスト生成済みだがペルソナ本文未保存の状態を明示しなければならない。

#### Scenario: npc_personas の属性が抽出元と一致する
- **WHEN** リクエスト生成前に `npc_personas` へ下書き保存する
- **THEN** `race` には NPC の種族値のみが保存されなければならない
- **AND** `sex` / `voice_type` / `source_plugin` は空でなく保存されなければならない（入力に値がある場合）

#### Scenario: source_plugin 欠損時はファイル名から補完される
- **WHEN** 入力データの `source_plugin` が空で、入力パスまたは入力名に `hoge.esm` / `hoge.esl` / `hoge.esp` が含まれる
- **THEN** システムは該当拡張子付きファイル名を `source_plugin` として保存しなければならない

#### Scenario: source_plugin を補完できない場合は UNKNOWN を設定する
- **WHEN** `source_plugin` が空で、入力情報から `*.esm|*.esl|*.esp` を抽出できない
- **THEN** システムは `source_plugin` に `UNKNOWN` を保存しなければならない

#### Scenario: npc_dialogues の editor_id はレスポンス欠損時にフォールバックされる
- **WHEN** dialogue response の `editor_id` が空で、dialogue group 側に `editor_id` がある
- **THEN** `npc_dialogues.editor_id` には group 側 `editor_id` を保存しなければならない

#### Scenario: npc_dialogues は原文中心で保存される
- **WHEN** ペルソナ用途の会話データを `npc_dialogues` に保存する
- **THEN** システムは `source_text` を保存しなければならない
- **AND** `translated_text` の更新を必須処理として扱ってはならない

#### Scenario: リクエスト生成時に status は draft になる
- **WHEN** `PreparePrompts` が `npc_personas` へ下書き保存と生成リクエスト保存を完了する
- **THEN** システムは対象行の `status` に `draft` を保存しなければならない
- **AND** `generation_request` が存在しても `status` を `generated` にしてはならない

### Requirement: ペルソナ保存は再開時も冪等でなければならない
MasterPersona の保存フェーズは、再試行または再開が発生しても同一 NPC を `source_plugin + speaker_id` で一意に識別し、重複レコードを作成せず、`overwrite_existing` に応じて更新または保持として確定保存しなければならない。保存が成功した行は `npc_personas.status` を英語値 `generated` に更新し、リクエスト生成時の `draft` と区別できなければならない。

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

#### Scenario: 保存成功時に status は generated になる
- **WHEN** `SavePersona` がペルソナ本文の保存を完了する
- **THEN** システムは対象行の `status` を `generated` に更新しなければならない
- **AND** `persona_text` が確定した行を `draft` のまま残してはならない

### Requirement: MasterPersona 一覧は保存済みダイアログ件数を表示しなければならない
MasterPersona のペルソナ一覧は、`npc_personas.dialogue_count` のような生成時スナップショット値ではなく、現在 `npc_dialogues` に保存されている関連会話件数を表示しなければならない。システムは一覧取得時に関連ダイアログを集計して表示用 DTO に反映し、同じ DTO に `npc_personas.status` の英語値 `draft` / `generated` を含めなければならない。フロントエンドはこの状態値を使って `下書き` / `生成済み` として表示しなければならない。

#### Scenario: 一覧件数は関連ダイアログ数から算出される
- **WHEN** ユーザーが MasterPersona のペルソナ一覧を開く
- **THEN** システムは各ペルソナについて `npc_dialogues` の関連件数を集計して返さなければならない
- **AND** `npc_personas.dialogue_count` を一覧表示の根拠として用いてはならない

#### Scenario: ダイアログ件数が更新されると一覧表示も追従する
- **WHEN** 既存ペルソナに紐づく `npc_dialogues` が追加または削除される
- **THEN** 次回一覧取得時のセリフ数は最新の関連件数を表示しなければならない

#### Scenario: 一覧は status に応じた表示名を出す
- **WHEN** 一覧取得結果の `status` が `draft` または `generated` を含む
- **THEN** フロントエンドは `draft` を `下書き`、`generated` を `生成済み` として表示しなければならない
- **AND** 全件を固定の完了表示にしてはならない

## ADDED Requirements

### Requirement: MasterPersona 一覧はステータスでフィルタできなければならない
MasterPersona 一覧は、既存の検索語とプラグイン絞り込みに加えて、`draft` / `generated` の状態で絞り込みできなければならない。フィルタ条件は一覧表示にのみ適用され、元の一覧データを破壊してはならない。

#### Scenario: 下書きだけを絞り込める
- **WHEN** ユーザーがステータスフィルタで `下書き` を選択する
- **THEN** 一覧には `status='draft'` のペルソナだけが表示されなければならない

#### Scenario: 生成済みだけを絞り込める
- **WHEN** ユーザーがステータスフィルタで `生成済み` を選択する
- **THEN** 一覧には `status='generated'` のペルソナだけが表示されなければならない

#### Scenario: ステータス解除で全件に戻せる
- **WHEN** ユーザーがステータスフィルタを未選択に戻す
- **THEN** システムは他の検索条件だけを維持しつつ、全ステータスの一覧を再表示しなければならない
