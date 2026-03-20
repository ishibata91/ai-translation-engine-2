# Logic Design

## Scenario
`データロード` phase で重複ファイルの読み込みをブロックし、`TerminologyPanel` では単語翻訳 phase の実行前に対象単語リストを表示できるようにする。

## Goal
- 同じプラグインを 2 重に処理しないため、データロード時に同一ファイル名の重複読み込みを防ぐ。
- translation flow workflow は terminology phase の実行 summary とは別に、artifact に保存済みの実行対象 preview を task 単位で返す。
- workflow は業務ルールの再導出ではなく、artifact から受け取った値の受け渡しと UI 向けページングに徹する。

## Responsibility Split
- Controller:
  `TaskController` は terminology summary と terminology target preview を Wails 向け API として公開する。
  taskID 解決だけを担い、対象抽出や表示整形ロジックは持たない。
- Workflow:
  translation flow workflow は taskID を受け取り、translationinput artifact から Terminology preview 用 DTO を投影する。
  summary 取得、preview 取得、phase 実行を orchestration する。
  `Variant` や `SourceFile` を業務ルールから再導出しない。
- Slice:
  `terminology` slice は翻訳 request 構築と保存ルールを保持する。
  preview のために slice API を増やさず、workflow は slice の代わりにルール判断を持たない。
- Artifact:
  `translationinput` artifact は Terminology 対象 entry 群の正本を保存・参照する。
  `TerminologyEntry.Variant` は artifact で保持する。
  `TerminologyEntry.SourceFile` はデータロード時点でファイル名へ正規化して保持する。
- Runtime:
  変更しない。LLM 実行と進捗通知に集中する。
- Gateway:
  変更しない。
- Foundation:
  Terminology 対象 REC の共有定数を提供する。workflow と terminology slice は同じ正本を参照する。

## Data Flow
- 入力:
  taskID、page、pageSize、選択ファイル一覧
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
    TaskID    string                        `json:"task_id"`
    Page      int                           `json:"page"`
    PageSize  int                           `json:"page_size"`
    TotalRows int                           `json:"total_rows"`
    Rows      []TerminologyTargetPreviewRow `json:"rows"`
}
```

補足:
- `Variant` は artifact が保持する `full` / `short` / `single` をそのまま返す。
- `PairKey` は UI 表示不要なので preview DTO に含めない。
- `SourceFile` は artifact が保持するファイル名だけを返す。
- `target_count` は summary / preview の両方から削除する。

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
1. データロード phase で UI は選択候補のファイル名を既存の選択済み一覧とロード済み一覧に照合する。
2. 同一ファイル名が見つかった場合、UI はその候補を追加せず、重複ブロックメッセージを表示する。
3. 重複していないファイルだけがロードされ、artifact に保存される。
4. artifact は terminology entry 保存時に `SourceFile` をファイル名へ正規化して保存する。
5. artifact は terminology entry 保存時に `Variant` を正本として保存する。
6. frontend が `TerminologyPanel` 表示時に `GetTranslationFlowTerminology(taskID)` を呼び、summary を取得する。
7. frontend は同時に `ListTranslationFlowTerminologyTargets(taskID, page, pageSize)` を呼び、対象単語 preview を取得する。
8. controller は taskID を解決して workflow へ委譲する。
9. workflow は translationinput artifact から Terminology 対象 entry 群を読み出す。
10. workflow は `Variant` と `SourceFile` を再導出せず、そのまま preview DTO へ写す。
11. frontend は summary card と list card を描画する。
12. ユーザーが実行すると workflow は同じ対象集合を使って terminology phase を開始する。
13. 実行後に frontend は summary を再取得し、preview は維持または再取得する。

## Key Branches
- データロード時の重複候補
  - 同一ファイル名なら追加しない。
- 対象 0 件
  - preview は `Rows = []` を返す。
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
- 非 NPC の重複統合は preview では行わず、一覧実装を複雑化させないため保存先 entry 単位で返す。
- 初期ページサイズは 50 件とする。
- 初期ソートは `RecordType`, `SourceText`, `EditorID` を基本とし、NPC の隣接表示は artifact 側の保持順または最小限の補助ソートで担保する。

## Persistence Boundary
- 新規永続化は不要。
- preview DTO は都度投影し、DB や task state に二重保存しない。
- terminology phase 実行後も preview 元データは translationinput artifact を正本にする。
- path はデータロード時点以降の terminology preview には不要なので保持しない。

## Side Effects
- `pkg/workflow/translation_flow.go` に preview DTO と workflow interface 追加が必要になる。
- `pkg/controller/task_controller.go` に preview API 追加が必要になる。
- frontend hook の state / actions / adapters / E2E fixture に preview payload が増える。
- 既存の `TerminologyPhaseSummary` から `targetCount` を削除する必要がある。
- データロード UI に同一ファイル名ブロック判定とメッセージ表示が必要になる。

## Risks
- 同一ファイル名で別内容のファイルもブロック対象になるが、今回は `同じプラグインを2重で処理しない` ことを優先する。
- 非 NPC を entry 単位で表示すると、同一訳共有の対象が複数行に見えるため、UI の補足文言が必要になる。
- `target_count` を削除するため、既存の summary card と adapter 参照をまとめて整理する必要がある。

## Follow-up For Review
- 同一ファイル名ブロックがデータロード仕様として十分か。
- workflow が `Variant` / `SourceFile` を pass-through するだけで責務境界が明確か。
- `TerminologyPhaseSummary` から `targetCount` を削除しても phase 進行判断に不足が出ないか。
