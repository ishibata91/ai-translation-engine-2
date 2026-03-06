## ADDED Requirements

### Requirement: MasterPersona LLM 設定は永続化・再読込できなければならない
`config` は MasterPersona 用 LLM 設定（provider/model/endpoint/apiKey/temperature/maxTokens）を namespace 管理で保存し、画面起動時に再読込できなければならない。

#### Scenario: 設定保存後に再起動しても復元される
- **WHEN** ユーザーが MasterPersona 画面で設定を保存した後にアプリを再起動する
- **THEN** 画面初期化時に保存済み設定が読み込まれ、入力欄へ反映されなければならない

#### Scenario: 未保存時は安全な既定値を返す
- **WHEN** MasterPersona 設定が未保存である
- **THEN** `config` は空値エラーを返さず、既定値または空初期値を返さなければならない

#### Scenario: apiKey はローカル用途として平文保存される
- **WHEN** ユーザーが MasterPersona の `apiKey` を保存する
- **THEN** `config` は暗号化を必須とせず、ローカル設定値として保存・再読込できなければならない
