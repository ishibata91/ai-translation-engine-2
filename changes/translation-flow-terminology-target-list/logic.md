# Logic Design

## Scenario
`TerminologyPanel` で単語翻訳 phase の実行前に対象単語リストを表示できるようにする。

## Goal
translation flow workflow が、Terminology phase の実行 summary とは別に、実行対象 preview を task 単位で返せるようにし、frontend が対象範囲を確認可能な UI を構成できる状態にする。

## Responsibility Split
- Controller:
  `TaskController` は terminology summary と terminology target preview を Wails 向け API として公開する。
  taskID 解決だけを担い、対象抽出や表示整形ロジックは持たない。
- Workflow:
  translation flow workflow は taskID を受け取り、translationinput artifact から Terminology preview 用 DTO を投影する。
  summary 取得、preview 取得、phase 実行を orchestration する。
- Slice:
  `terminology` slice は翻訳 request 構築と保存ルールを保持する。
  preview のために slice API を増やさず、workflow が既存の対象集合ルールに従って投影する。
- Artifact:
  `translationinput` artifact は Terminology 対象 entry 群の正本を保存・参照する。
  artifact 自体は UI 向けのページングや表示文言の判断を持たない。
- Runtime:
  変更しない。LLM 実行と進捗通知に集中する。
- Gateway:
  変更しない。
- Foundation:
  Terminology 対象 REC の共有定数を提供する。workflow と terminology slice は同じ正本を参照する。

## Data Flow
- 入力:
  taskID、page、pageSize
- 中間成果物:
  translationinput artifact に保存された Terminology 対象 entry 群
- 出力:
  frontend が描画する `TerminologyTargetPreviewPage`

## Preview DTO
UI が必要とする最小 DTO は以下とする。

```go
type TerminologyTargetPreviewRow struct {
    ID         string `json:"id"`
    RecordType string `json:"record_type"`
    EditorID   string `json:"editor_id"`
    SourceText string `json:"source_text"`
    Variant    string `json:"variant"`
    SourceFile string `json:"source_file"`
}

type TerminologyTargetPreviewPage struct {
    TaskID      string                        `json:"task_id"`
    Page        int                           `json:"page"`
    PageSize    int                           `json:"page_size"`
    TotalRows   int                           `json:"total_rows"`
    TargetCount int                           `json:"target_count"`
    Rows        []TerminologyTargetPreviewRow `json:"rows"`
}
```

補足:
- `Variant` は `full` / `short` / `single` を返す。
- `PairKey` は UI 表示不要なので preview DTO に含めない。
- `TargetCount` は list card の件数 badge と空 state 判定に使う。

## API Shape
summary と list は責務を分離する。

- 既存 API
  - `GetTranslationFlowTerminology(taskID)`
- 新規 API
  - `ListTranslationFlowTerminologyTargets(taskID, page, pageSize)`

理由:
- load phase でも `ListLoadedTranslationFlowFiles` と `ListTranslationFlowPreviewRows` が分離されている。
- preview にページングを持たせる以上、summary API に list 責務を混ぜない方が controller / workflow の境界が明確になる。
- frontend は summary state と preview state を別管理し、`状態を再読込` で両方を更新する。

## Main Path
1. frontend が `TerminologyPanel` 表示時に `GetTranslationFlowTerminology(taskID)` を呼び、summary を取得する。
2. frontend は同時に `ListTranslationFlowTerminologyTargets(taskID, page, pageSize)` を呼び、対象単語 preview を取得する。
3. controller は taskID を解決して workflow へ委譲する。
4. workflow は translationinput artifact から Terminology 対象 entry 群を読み出す。
5. workflow は Foundation の対象 REC 定義と terminology の variant ルールに従って preview DTO を構築する。
6. frontend は summary card と list card を描画する。
7. ユーザーが実行すると workflow は同じ対象集合を使って terminology phase を開始する。
8. 実行後に frontend は summary を再取得し、preview は維持または再取得する。

## Key Branches
- 対象 0 件
  - preview は `TargetCount = 0`、`Rows = []` を返す。
  - UI は空 state を表示し、実行ボタンを無効化する。
- preview 取得失敗
  - summary 取得成功は有効のままにし、list card だけエラー表示にする。
- 実行完了後の preview 再取得失敗
  - summary 完了状態は維持し、preview だけ再読込失敗として扱う。
- task 未解決
  - controller は task 解決失敗を返し、frontend は実行不能状態を維持する。

## Preview Projection Rules
- preview 対象は terminology phase が実際に使う対象集合と同一でなければならない。
- NPC FULL/SHRT は別行で返し、同一 NPC の組は並び順で隣接させる。
- 非 NPC の重複統合は preview では行わず、保存先 entry 単位で返す方針を第一候補にする。
理由:
  UI では `どの editor id が対象か` を確認したいので、request 単位へ潰すと追跡性が落ちる。
- 初期ソートは `RecordType`, `SourceText`, `EditorID` を基本とし、NPC は内部的に `PairKey`, `Variant` を補助キーに使って隣接表示を担保する。

## Persistence Boundary
- 新規永続化は不要。
- preview DTO は都度投影し、DB や task state に二重保存しない。
- terminology phase 実行後も preview 元データは translationinput artifact を正本にする。

## Side Effects
- `pkg/workflow/translation_flow.go` に preview DTO と workflow interface 追加が必要になる。
- `pkg/controller/task_controller.go` に preview API 追加が必要になる。
- frontend hook の state / actions / adapters / E2E fixture に preview payload が増える。

## Risks
- translationinput artifact が `SourceFile` や `Variant` を直接返せない場合、workflow 側で追加投影ロジックが必要になる。
- `summary.target_count` が request 数のままだと、preview 行数と食い違う可能性がある。
- 非 NPC を entry 単位で表示すると、同一訳共有の対象が複数行に見えるため、件数解釈を UI 文言で補う必要がある。

## Follow-up For Review
- `summary.target_count` を request 数ではなく preview 行数に寄せるべきか。
- preview 用 API を workflow interface に追加しても責務過多にならないか。
- Source File の表示粒度はファイル名だけで十分か。
