## MODIFIED Requirements

### Requirement: Master Persona のモデル設定は共有 LLM リクエスト設定 UI として再利用できなければならない
`frontend/src/components/ModelSettings.tsx` は、Master Persona 画面専用の実装に閉じず、他 feature でも同じ LLM リクエスト設定入力を再利用できる props 境界を提供しなければならない。provider、model、execution profile、temperature、endpoint、apiKey、contextLength、syncConcurrency などの入力は namespace と title を切り替えるだけで再利用できなければならない。

#### Scenario: terminology phase が ModelSettings を再利用する
- **WHEN** translation flow の単語翻訳 phase が LLM リクエスト設定を表示する
- **THEN** システムは `ModelSettings` を再利用して terminology 用設定を編集できなければならない
- **AND** master persona 固有の保存先や文言に固定してはならない

### Requirement: Master Persona の prompt 設定カードは共有 prompt editor として再利用できなければならない
`frontend/src/components/masterPersona/PromptSettingCard.tsx` は、Master Persona 画面専用の説明や文言に閉じず、他 feature でも system prompt / user prompt の編集 UI として再利用できなければならない。title、description、badge、補助文言、readOnly 状態は feature ごとに差し替えられなければならない。

#### Scenario: terminology phase が PromptSettingCard を再利用する
- **WHEN** translation flow の単語翻訳 phase が prompt 編集 UI を表示する
- **THEN** システムは `PromptSettingCard` を再利用して terminology 用 prompt を編集できなければならない
- **AND** master persona 固有の補助文言を固定表示してはならない
