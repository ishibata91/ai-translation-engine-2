# 辞書DB作成 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User as ユーザー
    participant UI as React UI
    participant Handler as Pipeline<br/>(HTTP Handler)
    participant Importer as DictionaryImporter<br/>(Interface)
    participant Parser as XMLParser
    participant Store as DictionaryStore<br/>(Interface)
    participant DB as Infrastructure<br/>(Shared *sql.DB)

    User->>UI: 複数のXMLファイルを選択
    User->>UI: 「インポート実行」をクリック
    UI->>Handler: POST /api/dictionary/import<br/>(multipart/form-data)
    
    loop 各XMLファイル
        Handler->>Importer: ImportXML(ctx, fileReader)
        Importer->>Parser: Parse(fileReader)
        Parser-->>Importer: []DictTerm
        
        Importer->>Store: SaveTerms(ctx, []DictTerm)
        Store->>DB: BEGIN TRANSACTION (using pooled connection)
        Store->>DB: INSERT/UPSERT INTO dictionaries (EDID, REC, Source, Dest, Addon)
        Store->>DB: COMMIT
        DB-->>Store: Success
        Store-->>Importer: Success
        Importer-->>Handler: ImportedCount
    end
    
    Handler-->>UI: 200 OK (TotalImportedCount)
    UI-->>User: インポート完了通知
```
