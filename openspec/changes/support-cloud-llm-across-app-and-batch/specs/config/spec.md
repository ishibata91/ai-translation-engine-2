## MODIFIED Requirements

### Requirement: Provider別設定の独立保存
同一画面/機能でプロバイダを切り替える設定は、`<base_namespace>.<provider>` に分離保存しなければならない。別プロバイダへの切替時に、他プロバイダ設定を上書きしてはならない。独立保存の対象には `model`、`endpoint`、`api_key`、`temperature`、`context_length` に加えて `bulk_strategy` を含めなければならない。

#### Scenario: MasterPersonaでプロバイダ別設定を保持する
- **WHEN** ユーザーが `master_persona.llm` で `lmstudio` と `gemini` を切り替えて各設定を保存する
- **THEN** `master_persona.llm.lmstudio` と `master_persona.llm.gemini` は独立して保持される
- **AND** 再度切り替えた際に、直前に保存した各プロバイダ固有値が復元される

#### Scenario: provider ごとに実行方式を保持する
- **WHEN** ユーザーが `gemini` を `batch`、`xai` を `sync` で保存してから provider を切り替える
- **THEN** `master_persona.llm.gemini.bulk_strategy` と `master_persona.llm.xai.bulk_strategy` は独立して保持されなければならない
- **AND** provider を再選択した際に対応する `bulk_strategy` が復元されなければならない

### Requirement: MasterPersona LLM 設定は永続化・再読込できなければならない
`config` は MasterPersona 用 LLM 設定（provider、model、endpoint、apiKey、temperature、contextLength、syncConcurrency、bulkStrategy）を namespace 管理で保存し、画面起動時に再読込できなければならない。`bulkStrategy` 未保存時は安全な既定値として `sync` を返さなければならない。

#### Scenario: 設定保存後に再起動しても復元される
- **WHEN** ユーザーが MasterPersona 画面で設定を保存した後にアプリを再起動する
- **THEN** 画面初期化時に保存済み設定が読み込まれ、入力欄へ反映されなければならない

#### Scenario: 未保存時は安全な既定値を返す
- **WHEN** MasterPersona 設定が未保存である
- **THEN** `config` は空値エラーを返さず、既定値または空初期値を返さなければならない

#### Scenario: apiKey はローカル用途として平文保存される
- **WHEN** ユーザーが MasterPersona の `apiKey` を保存する
- **THEN** `config` は暗号化を必須とせず、ローカル設定値として保存・再読込できなければならない

#### Scenario: bulkStrategy 未保存時は sync を返す
- **WHEN** 既存ユーザーの `master_persona.llm.<provider>` に `bulk_strategy` が存在しない
- **THEN** `config` は互換性を保つため `sync` を返さなければならない
- **AND** 既存設定だけで起動不能になってはならない
