# XMLエクスポート クラス図

## クラス構成

```mermaid
classDiagram
    class ExportInput {
        +PluginName string
        +SourceLanguage string
        +DestLanguage string
        +TermResults []ExportRecord
        +MainResults []ExportRecord
        +OutputFilePath string
    }

    class ExportRecord {
        +FormID string
        +EditorID string
        +RecordType string
        +SourceText string
        +TranslatedText string
    }

    class ExportGenerator {
        <<interface>>
        +GenerateXML(ctx context.Context, input ExportInput) error
    }

    class ExportGeneratorImpl {
        -merger ResultMerger
        -formatter XMLFormatter
        -writer FileWriter
        +GenerateXML(ctx context.Context, input ExportInput) error
    }

    class ResultMerger {
        <<interface>>
        +Merge(termResults []ExportRecord, mainResults []ExportRecord) []ExportRecord
    }

    class DefaultResultMerger {
        +Merge(termResults []ExportRecord, mainResults []ExportRecord) []ExportRecord
        -generateKey(record ExportRecord) string
    }

    class XMLFormatter {
        <<interface>>
        +Format(pluginName string, sourceLang string, destLang string, records []ExportRecord) ([]byte, error)
    }

    class SSTXMLFormatter {
        +Format(pluginName string, sourceLang string, destLang string, records []ExportRecord) ([]byte, error)
        -buildXMLDocument(...) SSTXMLRessources
        -normalizeRecordType(recType string) string
    }

    class FileWriter {
        <<interface>>
        +WriteFile(filePath string, data []byte) error
    }

    class StandardFileWriter {
        +WriteFile(filePath string, data []byte) error
    }

    %% XML Data Structures (encoding/xml)
    class SSTXMLRessources {
        +Params XMLParams
        +Content XMLContent
    }

    class XMLParams {
        +Addon string
        +Source string
        +Dest string
        +Version int
    }

    class XMLContent {
        +Strings []XMLString
    }

    class XMLString {
        +List int
        +SID string
        +Partial int
        +EDID string
        +REC string
        +Source string
        +Dest string
    }

    ExportInput --> ExportRecord
    ExportGenerator <|.. ExportGeneratorImpl : implements
    ExportGeneratorImpl --> ResultMerger : uses
    ExportGeneratorImpl --> XMLFormatter : uses
    ExportGeneratorImpl --> FileWriter : uses
    ResultMerger <|.. DefaultResultMerger : implements
    XMLFormatter <|.. SSTXMLFormatter : implements
    FileWriter <|.. StandardFileWriter : implements
    SSTXMLFormatter ..> SSTXMLRessources : creates
    SSTXMLRessources --> XMLParams
    SSTXMLRessources --> XMLContent
    XMLContent --> XMLString
```

## コンポーネントの説明
- **`ExportGenerator`**: エクスポート処理のエントリポイント。オーケストレーターから呼び出される。
- **`ResultMerger`**: 用語翻訳結果（Pass 1）と本文翻訳結果（Pass 2）を結合する。基本的には重複しない前提だが、設定ミスに備えて `EditorID` と `RecordType` の複合キーで重複チェックを行い、重複時は警告ログを出力しつつ後勝ちで上書きする防衛的ロジックを持つ。
- **`XMLFormatter`**: 統合されたレコードリストを `SSTXMLRessources` 構造体にマッピングし、`encoding/xml` パッケージを用いて XML バイト配列を生成する。
- **`FileWriter`**: 生成された XML バイト配列をファイルシステムに安全に（アトミックな書き込みなどを用いて）保存する。
