# 翻訳フロー 本文翻訳

翻訳フローの `本文翻訳` phase における preview 契約、phase 進行、resume / retry を定義する。

## Requirements

### Requirement: Translation Flow workflow は本文翻訳 phase を translation project task 配下で実行しなければならない
システムは、`TranslationFlow` の本文翻訳 phase を `translation_project` task 配下の phase として実行しなければならない。preview、翻訳実行、resume、retry は同じ `task_id` を使って追跡され、`ペルソナ生成` の後、`export` の前に位置付けられなければならない。

#### Scenario: 本文翻訳 phase が同一 task ID で追跡される
- **WHEN** workflow が本文翻訳 phase を開始または再表示する
- **THEN** システムは同じ `translation_project.task_id` を使って preview、実行、resume を追跡しなければならない
- **AND** 別 task を新規発行してはならない

### Requirement: workflow は frontend-only preview 契約から本文翻訳 state を hydrate しなければならない
システムは、backend 未実装の間も本文翻訳 phase を frontend-only 契約で hydrate しなければならない。workflow は translation target preview、phase 設定、行状態、draft 状態から本文翻訳 UI state を復元し、stale な UI state だけを根拠にしてはならない。

#### Scenario: reopen 時に本文翻訳 state が復元される
- **WHEN** ユーザーが既存 translation task を開き直して本文翻訳タブを開く
- **THEN** workflow は保存済みの translation phase 設定、summary、行状態、draft を使って phase state を復元しなければならない
- **AND** 未保存の一時表示フラグだけを根拠に復元してはならない

#### Scenario: reopen 時は選択カテゴリと選択行だけを復元する
- **WHEN** ユーザーが既存 translation task を reopen する
- **THEN** workflow は `選択カテゴリ` と `選択行` を復元しなければならない
- **AND** tree 開閉状態やスクロール位置を復元してはならない

### Requirement: workflow は preview 行を `会話` `クエスト` `その他` に分類しなければならない
システムは、本文翻訳 preview 行を `会話` `クエスト` `その他` の 3 区分へ分類しなければならない。`会話` は speaker 単位 dialogue 群、`クエスト` は quest ID を持つ record 群、`その他` はクエスト・会話以外の record 群として扱わなければならない。

#### Scenario: 同じ分類規則で一覧と詳細を構成する
- **WHEN** workflow が本文翻訳 preview を UI へ渡す
- **THEN** 同じ分類規則を使って一覧表示用ノードと詳細表示用メタデータを構成しなければならない
- **AND** 一覧側と詳細側でカテゴリ判定がずれてはならない
- **AND** `その他` はクエスト・会話以外の record を対象にしなければならない
- **AND** 参照コンテキストは、会話では `会話の属性` `NPC の属性` `ペルソナ` `参考単語`、クエストとその他では `本文の属性` `参考単語` のみを返さなければならない

### Requirement: workflow は translation phase 専用設定を独立して保存しなければならない
システムは、本文翻訳 phase のモデル設定と user prompt を `translation_flow.translation` namespace として独立保存しなければならない。system prompt は保存対象に含めず、category / recordType から都度導出しなければならない。

#### Scenario: user prompt 変更が translation phase に閉じる
- **WHEN** ユーザーが本文翻訳 phase の user prompt を変更する
- **THEN** workflow は変更を `translation_flow.translation` に保存しなければならない
- **AND** persona phase や他 phase の prompt 設定へ影響させてはならない

### Requirement: workflow は本文状態を `未翻訳` `AI翻訳済み` `確定` の 3 種で管理しなければならない
システムは、本文ごとの状態を `未翻訳` `AI翻訳済み` `確定` で一貫して管理しなければならない。翻訳成功時は `AI翻訳済み` とし、手修正内容は本文単位の `確定` 操作で保存したときのみ `確定` へ遷移させなければならない。

#### Scenario: 手修正は本文単位の確定操作で保存される
- **WHEN** ユーザーが `AI翻訳済み` 本文を編集する
- **THEN** workflow は自動保存してはならない
- **AND** 本文単位の `確定` 操作を受けた時点でだけ保存し、状態を `確定` に更新しなければならない

#### Scenario: `確定` 行の再編集では一覧ラベルを維持する
- **WHEN** ユーザーが `確定` 行の本文を再編集する
- **THEN** workflow は一覧ラベルを `確定` のまま維持しなければならない
- **AND** `取り消し` 操作でのみ確定解除できる状態を返さなければならない
- **AND** `取り消し` 押下時は未保存編集を破棄し、保存済みの確定本文を返さなければならない

### Requirement: workflow は translating 中に一覧操作と phase 遷移をロックしなければならない
システムは、本文翻訳実行中に一覧選択、カテゴリ切替、フィルタ、モデル設定変更、phase 遷移をロックしなければならない。workflow は UI のロック状態と `次へ` 可否を一貫して返さなければならない。

#### Scenario: translating 中は進行条件が固定される
- **WHEN** workflow が本文翻訳実行を開始する
- **THEN** workflow は `translating` state と lock 対象一式を返さなければならない
- **AND** `次へ` を無効にしなければならない

### Requirement: workflow は未確定編集がある状態で対象変更や移動を抑止しなければならない
システムは、未確定編集を持つ状態で他行選択、カテゴリ切替、検索、ページ移動、phase 遷移が要求された場合、警告モーダルを返さなければならない。ユーザーが破棄を明示した場合だけ遷移を継続できる。

#### Scenario: 未確定編集で警告モーダルを返す
- **WHEN** workflow が未確定編集を持つ状態で遷移要求を受け取る
- **THEN** workflow は警告モーダル表示に必要な状態を返さなければならない
- **AND** 破棄確認前に遷移を実行してはならない

### Requirement: workflow は partial failed 後に failed 行だけを retry しなければならない
システムは、本文翻訳結果の一部が失敗した場合、`AI翻訳済み` 行と `確定` 行を保持したまま、failed 行だけを retry 対象として扱わなければならない。retry は既存成功行を巻き戻してはならない。

#### Scenario: partial failed で failed 行だけを再試行する
- **WHEN** 本文翻訳 phase が `partialFailed` の後に再試行される
- **THEN** workflow は failed 行だけを retry 対象として返さなければならない
- **AND** `AI翻訳済み` 行と `確定` 行を未翻訳状態へ戻してはならない

#### Scenario: partial failed でも後続 phase へ進行できる
- **WHEN** failed 行が残ったまま本文翻訳 phase を終了する
- **THEN** workflow は failed 件数を summary に残したまま `次へ` を有効にできなければならない
- **AND** 後続 phase へ未翻訳行が残る事実を UI に示せる summary を返さなければならない

### Requirement: workflow は `次へ` 押下時に `未翻訳` 残件を警告しなければならない
システムは、`次へ` 押下時に `未翻訳` 行が残る場合、警告モーダルを返さなければならない。`AI翻訳済み` と `確定` は警告対象に含めてはならない。

#### Scenario: `未翻訳` 残件がある場合だけ警告モーダルを返す
- **WHEN** ユーザーが本文翻訳 phase で `次へ` を押す
- **THEN** workflow は `未翻訳` 行が残っている場合にのみ警告モーダル状態を返さなければならない
- **AND** 警告モーダルには未翻訳件数だけを含め、カテゴリ内訳を含めてはならない
- **AND** `AI翻訳済み` と `確定` のみで構成される場合は警告モーダルを返してはならない

### Requirement: workflow は本文翻訳対象 0 件のとき no-op 完了しなければならない
システムは、本文翻訳対象が 0 件である場合、runtime を要求せずに本文翻訳 phase を no-op 完了として扱わなければならない。summary は空対象を示し、失敗扱いにしてはならない。

#### Scenario: 対象 0 件では即時完了する
- **WHEN** workflow が本文翻訳 preview を構築した結果、対象行が 0 件である
- **THEN** workflow は本文翻訳 phase を empty として返さなければならない
- **AND** `次へ` を有効化しなければならない
- **AND** 実行中 state や retry state を返してはならない
