## 1. Contextual Domain Models Split
<!-- Models.goを各コンテキストごとに分割する -->

- [x] 1.1 `pkg/domain/models/base.go` を作成し、`BaseExtractedRecord` と `ExtractedData` を移動する
- [x] 1.2 `pkg/domain/models/dialogue.go` を作成し、`DialogueResponse` と `DialogueGroup` を移動する
- [x] 1.3 `pkg/domain/models/quest.go` を作成し、`Quest`, `QuestStage`, `QuestObjective` を移動する
- [x] 1.4 `pkg/domain/models/entity.go` を作成し、`NPC`, `Item`, `Magic`（および `IsFemale` メソッド）を移動する
- [x] 1.5 `pkg/domain/models/system.go` を作成し、`Location`, `Message`, `SystemRecord`, `LoadScreen` を移動する
- [x] 1.6 元の `pkg/domain/models/models.go` を削除し、ビルドエラーがないか確認する

## 2. Loader Slice Architecture
<!-- LoaderをDIで構成されるVertical Sliceにする -->

- [x] 2.1 `pkg/loader/provider.go` を作成し、`ProvideLoader() contract.Loader` を実装する
- [x] 2.2 `pkg/loader/loader.go`（および関連ファイルにある内部構造体）を、Interfaceの背後に隠蔽するリファクタリングを行う
- [x] 2.3 `cmd` または呼び出し元の Wire セット（`wire.NewSet`）に `loader.ProvideLoader` を追加する

## 3. Verification
<!-- 変更の動作確認 -->

- [x] 3.1 既存の Loader 関連のユニットテストが今まで通りPassすることを確認する
- [x] 3.2 既存のJSONファイルをパースする結合テスト（またはCLI実行）を行い、パース結果に差分がないことを確認する
