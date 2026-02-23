## Why

現在の Skyrim Mod 翻訳エンジン (v2.0) の仕様では、最終的な出力が JSON 形式のみに定義されており、xTranslator が要求する XML 形式（SSTXMLRessources）へのエクスポート仕様が存在しません。
また、XML 生成に必須となる `<EDID>`、`<REC>` (シグネチャ)、`sID` などのデータが、各スライス（Loader, ProcessManager, Pass2Translator）を通過する過程で欠落または変形してしまうという仕様上の問題があります。
本提案は、これらの問題を解決し、データが抽出からXML出力まで正確に伝播し、xTranslator互換のXMLを出力するアーキテクチャを定義することを目的とします。

## What Changes

- JSON を読み込み、xTranslator 形式の XML を出力する新しい垂直スライス Export Slice を新設します。
- DSD形式JSONでの出力仕様を廃棄します。
- requirements.md に xTranslator互換のXML形式(SSTXMLRessources)を出力する Export Slice を備える旨を追記します。
- Pass2TranslatorSlice において、JSONの `type` フィールドの値を丸めずシグネチャを含めたフル形式で保存するよう仕様を変更します。
- ProcessManagerSlice において、階層構造を持つレコードを展開して TranslationRequest にマッピングする際、親レコードの ID (FormID) と EditorID を子要素に継承させる仕様を追加します。
- LoaderSlice (および抽出スクリプト extractData.pas) において、QuestStageとQuestObjectiveのDTOで親のIDとEditorIDを保持する拡張を行います。また、抽出時のバグを防ぐため、stage_indexとlog_indexの分離、およびorderの明示的出力の要件を追加します。

## Capabilities

### New Capabilities
- `ExportSlice`: Pass 2が出力したJSONを読み込み、xTranslator互換のXML形式(SSTXMLRessources)を出力する機能。

### Modified Capabilities
- `requirements`: 全体要件として、XMLエクスポートの責務を追加。
- `Pass2TranslatorSlice`: 翻訳結果出力時、type(シグネチャ)のフォーマットを維持するよう要件を変更。
- `ProcessManagerSlice`: リクエスト生成時、親レコードのID/EditorIDの継承要件を追加。
- `LoaderSlice`: DTOの拡張および、データ抽出時のバグ修正（インデックス分離と順序保証）要件を追加。

## Impact

- 影響を受ける仕様ドキュメント:
  - `specs/ExportSlice/spec.md` (新規)
  - `specs/requirements.md`
  - `specs/Pass2TranslatorSlice/spec.md`
  - `specs/ProcessManagerSlice/spec.md`
  - `specs/LoaderSlice/spec.md`
- 影響を受けるコンポーネント: 抽出スクリプト(`extractData.pas`), `LoaderSlice` (DTO), `ProcessManagerSlice`, `Pass2TranslatorSlice`, および新規の `ExportSlice`
