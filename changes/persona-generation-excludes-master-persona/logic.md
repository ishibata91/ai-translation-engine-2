# Logic Design

## Scenario
翻訳フローの `ペルソナ生成` phase が、translation input artifact から検出した NPC 候補を `source_plugin + speaker_id` で正規化し、既存 Master Persona を持つ候補を除外したうえで、未保有 NPC だけを生成して final 成果物へ保存する。

## Goal
翻訳プロジェクトが、既存 Master Persona を正本として再利用しながら不足分だけを補完生成し、preview・execute・resume・retry のすべてで同じ候補集合と除外ルールを保てるようにすること。

## Responsibility Split
- Controller:
  - `ListTranslationFlowPersonaTargets` `RunTranslationFlowPersonaPhase` `GetTranslationFlowPersonaPhase` の外部入力境界を提供する。
  - `task_id` の受け渡し、request/prompt DTO 整形、workflow 呼び出しだけを行い、Master Persona lookup や候補除外を持たない。
- Workflow:
  - translation input artifact から persona phase 用の NPC / dialogue 入力を読み出し、translation project task 境界へ束ねる。
  - preview と execute の両方で同じ persona slice planner を呼び、`検出数` `再利用数` `新規生成数` `失敗数` を持つ phase summary を管理する。
  - `新規生成数 == 0` の場合は runtime を呼ばずに no-op 完了を返す。
  - retry / resume では `生成失敗` または未生成行だけを再 dispatch し、`既存 Master Persona` と `生成済み` は再送しない。
  - `translation_flow.persona.progress` 相当の phase 進捗を通知する。
- Slice:
  - persona slice が候補正規化、`source_plugin + speaker_id` の lookup key 正規化、既存 final lookup、dialogue 収集、prompt 構築、生成結果保存を一貫して担う。
  - preview 用候補一覧と execute 用 request 計画を同じ内部 planner から導出し、候補集合のずれを許さない。
  - `既存 Master Persona` 行は request 化せず、final 成果物の persona 本文を詳細表示用に返せるようにする。
  - 生成成功した行は `master_persona_artifact` final に保存し、同一 task の再表示時に `生成済み` として復元できるようにする。
- Artifact:
  - `translationinput` artifact は translation project task 配下の NPC、dialogue、source file 情報を正本として保持する。
  - `master_persona_artifact` は `source_plugin + speaker_id` をキーにした final persona の正本であり、除外判定と詳細表示の根拠になる。
  - task 単位の persona phase summary と row state は translation flow 側の task ストアに保持し、final 成果物とは分離する。
  - persona slice の一時 request / dialogue handoff は task 単位の temp artifact に保持し、queue や UI state を正本にしない。
- Runtime:
  - persona slice / workflow が作った `新規生成対象` request だけを LLM へ dispatch する。
  - `既存 Master Persona` 行の lookup や候補除外、phase 完了判定は持たない。
- Gateway:
  - LLM 実行と datastore 技術接続を提供する。
  - lookup key 正規化、既存 persona 除外、retry 対象判定などの業務ルールを持たない。

## Data Flow
- 入力
  - `task_id`
  - translation input artifact に保存された NPC / dialogue / source file 情報
  - persona phase の request config と prompt config
  - `master_persona_artifact.LookupKey`
- 中間成果物
  - 正規化済み persona candidate 集合
  - `existing_final` / `pending_generation` / `generated` / `failed` の row state
  - pending candidate だけで作る `[]llmio.Request`
  - phase summary (`detected_count`, `reused_count`, `pending_count`, `generated_count`, `failed_count`)
  - task 単位の temp persona artifact
- 出力
  - persona target preview page
  - persona phase summary
  - `master_persona_artifact` final に保存された generated persona
  - 後続 phase が参照できる persona availability 状態

## Main Path
1. Workflow が translation input artifact から persona phase 用の NPC / dialogue 入力を取得する。
2. Workflow が入力を persona slice planner に渡し、`source_plugin + speaker_id` を正規化した candidate 集合を作らせる。
3. persona slice が各 candidate について `master_persona_artifact.LookupKey` で final 成果物を検索し、`existing_final` と `pending_generation` に分割する。
4. Workflow が preview 用 DTO と phase summary を作り、UI へ `検出数` `再利用数` `新規生成数` を返す。
5. 実行時、Workflow は同じ planner を再度呼び、`pending_generation` だけで request 計画を作る。
6. `pending_generation` が 0 件なら、Workflow は no-op 完了 summary を保存して runtime を呼ばずに返す。
7. `pending_generation` が 1 件以上なら、Workflow は runtime へ request を渡し、実行結果を persona slice `SaveResults` に渡す。
8. persona slice は成功結果を `master_persona_artifact` final へ保存し、task 単位 summary / row state を更新する。
9. Workflow は `generated_count` `failed_count` `progress` を更新し、UI へ最終 summary を返す。
10. UI は `既存 Master Persona` と `生成済み` をまとめて後続 phase へ handoff する。

## Key Branches
- 全件再利用:
  - candidate 全件が `existing_final` の場合、Workflow は no-op 完了 summary を返す。
  - Runtime と LLM は呼ばれない。
- 対象 0 件:
  - translation input artifact から persona candidate が 0 件なら empty summary を返し、phase は skip 可能状態になる。
- 同一候補の重複:
  - 同じ `source_plugin + speaker_id` を持つ candidate は 1 件へ統合し、source file や dialogue は同一候補へ束ねる。
  - preview と execute で異なる統合結果を返してはならない。
- partialFailed:
  - 一部 request だけ失敗した場合、成功済み行は final に保存したまま `failed` 行だけを row state に残す。
  - Workflow は `次へ` を許可しつつ `再試行` で failed 行だけを再 dispatch できる summary を返す。
- resume / retry / cancel:
  - resume は task summary と final 成果物を基に row state を再構成する。
  - retry は failed または未生成行だけを再計画し、`existing_final` / `generated` を再送しない。
  - cancel は本 change では追加しない。
- source plugin 正規化:
  - lookup key に必要な `source_plugin` が入力から欠ける場合、persona slice は既存の正規化規則に従って source file 名または `UNKNOWN` を使う。
  - preview と execute で異なる補完値を使ってはならない。

## Persistence Boundary
- `translationinput` artifact:
  - translation project task に紐づく raw NPC / dialogue / source file 情報
  - persona phase summary と row state の task 単位保存先
- `master_persona_artifact` final:
  - `source_plugin + speaker_id` をキーにした generated persona の正本
  - 既存 Master Persona の除外判定元と詳細表示の参照先
- task-scoped temp artifact:
  - request 構築中の dialogue や generation request の一時保存
  - final 成果物へ昇格する前の中間状態だけを保持する
- UI state:
  - 選択行や一時表示 state だけを持ち、除外判定や final persona の正本になってはならない

## Side Effects
- translation input artifact から NPC / dialogue を読み出す
- `master_persona_artifact` final を lookup する
- pending candidate だけを LLM 実行する
- 生成成功結果を `master_persona_artifact` final へ保存する
- translation project task の persona phase summary を更新する
- progress notifier で `persona` phase の進捗を通知する

## Risks
- preview と execute が別経路で candidate 集合を作ると、`新規生成数` と request 件数がずれる
- lookup key 正規化が master persona 側と一致しないと、既存 final を持つ NPC を誤って再生成する
- row state を final 成果物より優先してしまうと、resume 時に stale な除外判定を使う
- `partialFailed` で `次へ` を許可するため、後続 phase が persona 欠損行を許容する前提を実装で明示する必要がある
- translation input artifact が persona phase 用の NPC / dialogue 取得 API をまだ持たないため、artifact contract の追加が必要になる

## Context Board Entry
```md
### Logic Design Handoff
- 確定した責務境界: workflow は phase summary と dispatch 制御、persona slice は candidate 正規化と既存 Master Persona 除外と保存、artifact は raw input / final persona / task summary の保存境界を担う
- docs 昇格候補: translation flow persona phase の新規正本 docs、既存 Master Persona 除外規則、all-cached no-op 完了、duplicate candidate の lookup key 統合、failed 行だけの retry 規則
- review で見たい論点: preview/execute の candidate planner 共通化、task summary と final artifact の優先関係、partialFailed 時の次 phase 続行条件
- 未確定事項: なし
```
