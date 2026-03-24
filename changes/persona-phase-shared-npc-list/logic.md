# Logic Design

## Scenario
MasterPersona と TranslationFlow Persona Phase の shared NPC list

## Goal
一覧コンポーネントだけを共有し、MasterPersona の final 成果物表示と TranslationFlow の phase preview / runtime state 表示を無理なく両立させる。

## Responsibility Split
- Controller: 既存の `PersonaController.ListNPCs` と `TaskController.ListTranslationFlowPersonaTargets` を維持し、一覧取得の入口を変えない。必要なら既存 payload に含まれる `editor_id` を frontend へ露出する。
- Workflow: `TranslationFlowService` は row state と page 単位 preview を決定し続け、shared component の存在を知らない。
- Slice: persona slice は final artifact 由来の MasterPersona 一覧 DTO を返し続ける。
- Artifact: `master_persona_artifact` は MasterPersona 一覧の正本のまま据え置き、translation task preview の正本にしない。
- Runtime: request state と retry / resume の追跡を継続し、一覧共有のための追加責務を持たない。
- Gateway: 変更なし。

## Data Flow
- 入力
  - MasterPersona: `PersonaNPCView` と `PersonaDialogueView`
  - TranslationFlow: `PersonaTargetPreviewPage` と `PersonaPhaseResult`
- 中間成果物
  - frontend 専用の shared row contract
  - row ごとの stable selection key
- 出力
  - shared list component への rows / page / selection props
  - 親 component への row select / page change callback

## Main Path
- `frontend` に shared persona list row contract を定義し、`formId`, `sourcePlugin`, `npcName`, `editorId`, `updatedAt`, `stateLabel`, `stateTone` のような共通表示値へ正規化する。
- shared list component は `DataTable` を内包し、共通ヘッダー、共通列順、共通ページャー、共通選択スタイルだけを持つ。
- `MasterPersona` は既存の検索・プラグイン絞り込み・選択状態を維持したまま shared list へ rows を渡す。
- `TranslationFlow PersonaPanel` は `speakerId` を `formId` に写像し、phase state を補助表示用 props として shared list へ渡す。
- 右側の詳細ペインは各画面の既存 component を維持し、shared list は選択結果だけを返す。

## Key Branches
- TranslationFlow の `loadingTargets` / `empty` / `failed` は phase カード側で管理し、shared list は rows の有無と loading 表示に集中する。
- `editor_id` を MasterPersona 一覧にも表示したい場合、backend 変更ではなく既存 payload の frontend 正規化だけで吸収する。
- TranslationFlow の検索 / フィルタを shared list に含める場合は、別 change で API 契約を追加検討する。

## Persistence Boundary
永続化の追加変更は行わない。shared list のために新しい artifact や task metadata は作らない。

## Side Effects
- frontend component 分割
- shared row mapper の追加
- 既存 page component の props 再配線

## Risks
- detail pane まで共有しようとすると、MasterPersona だけが持つ `generationRequest` や `updatedAt` に TranslationFlow が引きずられる
- row key の決め方が画面ごとにぶれると selection 復元が壊れる
- TranslationFlow でフィルタ UI まで同一化すると backend の page/filter 契約拡張が必要になる

## Context Board Entry
```md
### Logic Design Handoff
- 確定した責務境界: 共有対象は list component のみ、row 正規化は親 hook、row state 判定は workflow
- docs 昇格候補: TranslationFlow persona 一覧は MasterPersona と同じ table shell と選択挙動を使う
- review で見たい論点: shared scope を list-only に留められているか、row state 可視性が落ちていないか
- 未確定事項: TranslationFlow 側の検索 / フィルタ戦略
```
