## Why

translation flow から `master persona` と `Dictionary Builder` の成果物を逐次参照したいが、現在は両者の成果物が各 slice ローカル DB に閉じており、slice 間 handoff の共有境界になっていない。`artifact` を共有成果物の正本にしつつ、workflow は artifact を直接読まず slice 経由で扱う形へ揃える必要がある。

## What Changes

- `Dictionary Builder` の共有成果物を `artifact` 側へ移し、translation flow から slice 経由で参照可能な正本ストアとして定義する。
- `MasterPersona` の共有成果物を `artifact` 側へ移し、translation flow と画面表示が同じ最終成果物を slice 経由で参照する正本ストアとして定義する。
- `MasterPersona` 一覧は下書き取得をやめ、生成済み成果物だけを全取得する形へ変更する。
- `MasterPersona` 一覧からステータス表示、ステータス由来の分岐、セリフ数表示を取り除く。
- `MasterPersona` の task 終了時に、下書きやリクエスト準備用の中間生成物を cleanup する。
- `pkg/slice/dictionary` と `pkg/slice/persona` の保存契約のうち、共有成果物に該当する部分を `pkg/artifact` へ移す。
- 各 artifact について、保存対象データ、cleanup ルール、slice からの参照境界を定義する。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `slice/dictionary`: Dictionary Builder の共有辞書成果物を artifact 正本へ移し、slice 経由で保存・更新・検索する要件へ変更する。
- `persona`: MasterPersona の最終ペルソナ成果物と task 単位の中間生成物を artifact で管理し、一覧は生成済み成果物のみを表示し、下書き取得、ステータス表示、セリフ数表示を持たないよう要件を変更する。
- `workflow/master-persona-execution-flow`: MasterPersona は task 終了時に task 単位の中間生成物を cleanup しなければならないよう要件を変更する。

## Impact

- `pkg/artifact/dictionary_artifact` と `pkg/artifact/master_persona_artifact` に保存・検索契約、構造化テーブル、migration が追加される。
- `pkg/slice/dictionary` と `pkg/slice/persona` は共有成果物の保存先として artifact 契約を使う形へ変わる。
- `pkg/workflow` は artifact を直接参照せず、slice 契約だけを束ねる前提へ揃える必要がある。
- `MasterPersona` UI は最終成果物の全取得に一本化され、status 関連 UI とセリフ数表示を削除する必要がある。
- `openspec/specs/governance/database-erd/spec.md` には shared artifact ストアを正本として扱う新規テーブルと、中間生成物 cleanup の責務境界を追記する必要がある。
