# XMLエクスポート (Export Slice) 仕様書

## 概要
Term Translator Slice (Pass 1) の用語翻訳結果と、Pass 2 Translator Slice (Pass 2) の本文翻訳結果を統合し、xTranslator エディタで読み込み可能なXML形式（SSTXMLRessources）として出力する機能である。
これにより、AIによる翻訳結果をxTranslator経由でSkyrimのプラグイン（.esp / .esm）に適用可能とする。

当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則を捨て、**本Slice自身が「XMLマッピングロジック」「独自DTO定義」「ファイル出力ロジック」の全ての責務を負う。** 外部機能のデータモデルには一切依存せず、単一の明確なコンテキストとして自己完結する。

## 背景・動機
- Skyrimの翻訳エコシステムにおいて、xTranslatorはデファクトスタンダードの翻訳適用ツールである。
- 本システムで生成された高品質なAI翻訳結果をゲームに反映させるためには、xTranslator互換のXML形式で出力する必要がある。
- 用語翻訳（名詞）と本文翻訳（会話・書籍等）は別のスライスで処理・保存されるため、最終的にこれらを1つのエクスポートファイルに統合する役割が必要である。

## スコープ
### 本Sliceが担う責務
1. **翻訳結果の統合**: ProcessManagerから渡される用語翻訳結果と本文翻訳結果のリストを受け取り、統合する。
2. **XMLフォーマットの構築**: 統合されたデータを元に、`xTranslator XML (SSTXMLRessources)` 形式のデータ構造を構築する。
3. **ファイルへの出力**: 構築したXMLデータを指定されたパスにファイルとして書き出す。

### 本Sliceの責務外
- LLMによる翻訳実行（Term Translator / Pass 2 Translatorの責務）
- 元データの抽出（Loader Sliceの責務）
- DBからのデータ読み出し（オーケストレーター/Process Managerがデータを収集して本Sliceに渡す）

## 要件

### 1. 独立性: エクスポート用データの受け取りと独自DTO定義
**Reason**: スライスの完全独立性を確保するAnti-Corruption Layerパターンを適用し、他スライスのDTOへの依存を排除するため。
**Migration**: Process Managerが各DBやJSONから結果を読み集め、本Slice専用の `ExportInput` DTOにマッピングして渡す。

#### Scenario: 独自定義DTOによる初期化とエクスポート処理
- **WHEN** オーケストレーター層から本スライス専用の入力DTO（`ExportInput`）が提供された場合
- **THEN** 外部パッケージのDTOに一切依存することなく、XMLファイルの構築と保存を完結できること
- **AND** `specs/refactoring_strategy.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

### 2. データ構造

**`ExportInput` 構造体**:
```go
type ExportInput struct {
    PluginName      string          // 対象プラグイン名 (例: "Dawnguard.esm")
    SourceLanguage  string          // 原文言語 (例: "english")
    DestLanguage    string          // 翻訳言語 (例: "japanese")
    TermResults     []ExportRecord  // Pass 1: 用語翻訳結果
    MainResults     []ExportRecord  // Pass 2: 本文翻訳結果
    OutputFilePath  string          // 出力先のファイルパス
}
```

**`ExportRecord` 構造体**:
```go
type ExportRecord struct {
    FormID         string  // フォームID (オプション)
    EditorID       string  // エディターID
    RecordType     string  // レコードタイプ (例: "WEAP:FULL", "INFO:NAM1")
    SourceText     string  // 英語原文
    TranslatedText string  // 日本語訳
}
```

### 3. 翻訳結果の統合 (Data Merge)
Pass 1（Term Translator）と Pass 2（Pass 2 Translator）は、処理対象のレコードタイプが完全に分かれている（例: Pass 1は `WEAP:FULL`、Pass 2は `INFO:NAM1` など）ため、**正常な設定下において `TermResults` と `MainResults` の間で同一レコード（`EditorID` と `RecordType` のペアが一致）が競合することはない**。

**統合ルール**:
1. `TermResults` と `MainResults` を結合（Append）する。
2. 万が一、設定ミス等により重複するレコードが検出された場合は、構造化ログとして「重複レコードの警告（Warning）」を出力し、後勝ち（通常は後続の `MainResults`）で上書きする防衛的実装とする。

### 4. xTranslator XML フォーマット要件
`specs/xtranslator_xml_spec.md` に基づき、以下のXMLを出力する。

**XML宣言**:
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
```

**ルート要素**:
`<SSTXMLRessources>`

**メタデータ (`<Params>`)**:
- `<Addon>`: `ExportInput.PluginName` の値
- `<Source>`: `ExportInput.SourceLanguage` の値
- `<Dest>`: `ExportInput.DestLanguage` の値
- `<Version>`: `2` (固定)

**翻訳データ (`<Content>` / `<String>`)**:
各レコードは `<Content>` 配下に `<String>` 要素として列挙する。
- 属性 `List="0"` (固定)
- 属性 `sID`: 元データの ID (例: `0x001234|Skyrim.esm`) から **8桁の16進数** 部分 (例: `00001234`) を抽出し、大文字で設定する
- 属性 `Partial="1"` (固定)
- 子要素 `<EDID>`: `ExportRecord.EditorID` (タグマッピングルール)
- 子要素 `<REC>`: `ExportRecord.RecordType` を正規化して設定 (例: `INFO NAM1` -> `INFO:NAM1`)
- 子要素 `<Source>`: `ExportRecord.SourceText` (XMLエスケープ処理)
- 子要素 `<Dest>`: `ExportRecord.TranslatedText` (XMLエスケープ処理)

### 5. XMLエスケープ処理
- 抽出テキストには `<` `>` `&` `"` `'` などの予約文字が含まれる。
- 出力時に標準のXMLエスケープ処理（Goの `encoding/xml` による処理など）を必ず行い、整形式（Well-formed）なXMLを生成すること。
- CDATAセクションは使用せず、エスケープ文字に変換する（例: `<` → `<`）。

### 6. ファイル出力
- 指定された `OutputFilePath` に対して、UTF-8エンコーディングでファイルを書き出す。
- ファイル書き込みはアトミックに行うか、エラー時にファイルが破損した状態で残らないようにハンドリングする（一時ファイルに書き出してからリネームする等）。

### 7. 進捗と結果通知
- 出力処理は高速に完了するため、途中のコールバック通知は不要とする。
- 成功時は `nil` を返し、エラー時は詳細なエラーメッセージを返す。

### 8. メインインターフェース

**`ExportGenerator` インターフェース**:
```go
// ExportGenerator は翻訳結果をXMLファイルとして出力する
type ExportGenerator interface {
    GenerateXML(ctx context.Context, input ExportInput) error
}
```

### 9. ライブラリの選定
- XML生成: Go標準 `encoding/xml`
- ファイルI/O: Go標準 `os`, `io`

## 関連ドキュメント
- [クラス図](./export_slice_class_diagram.md)
- [シーケンス図](./export_slice_sequence_diagram.md)
- [テスト設計](./export_slice_test_spec.md)
- [xTranslator XML仕様](../xtranslator_xml_spec.md)

---

## ログ出力・テスト共通規約

> 本スライスは `refactoring_strategy.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
