## ADDED Requirements

### Requirement: MasterPersona Prompt 設定は永続化・再読込できなければならない
`config` は MasterPersona 用 prompt 設定を `master_persona.prompt` namespace で保存し、画面起動時に再読込できなければならない。少なくとも `user_prompt` と `system_prompt` を個別キーで保持し、未保存時は空値エラーではなく既定値を返さなければならない。

#### Scenario: Prompt 設定保存後に再起動しても復元される
- **WHEN** ユーザーが MasterPersona 画面で `user_prompt` を編集した状態で画面を離脱またはアプリを再起動する
- **THEN** `config` は `master_persona.prompt.user_prompt` を保存し、次回表示時に同じ内容を返さなければならない
- **AND** `system_prompt` も同一 namespace から再読込できなければならない

#### Scenario: Prompt 設定が未保存でも既定値を返す
- **WHEN** `master_persona.prompt` namespace に保存値が存在しない
- **THEN** `config` は空値エラーを返さず、MasterPersona が表示可能な既定の `user_prompt` と `system_prompt` を返さなければならない
