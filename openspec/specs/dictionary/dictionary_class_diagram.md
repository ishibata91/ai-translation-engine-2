# 辞書DB作成 クラス図

```mermaid
classDiagram
    class DictTerm {
        +String EDID
        +String REC
        +String Source
        +String Dest
        +String Addon
    }

    class DictionaryImporter {
        <<Interface>>
        +ImportXML(ctx context.Context, file io.Reader) (int, error)
    }

    class DictionaryStore {
        <<Interface>>
        +SaveTerms(ctx context.Context, terms []DictTerm) error
    }

    class DictionaryService {
        -DictionaryStore store
        -DictionaryImporter importer
        +GetSources(ctx context.Context) ([]DictSource, error)
        +GetEntries(ctx context.Context, sourceID int64) ([]DictTerm, error)
        +StartImport(ctx context.Context, filePath string) (int64, error)
        +UpdateEntry(ctx context.Context, term DictTerm) error
        +DeleteEntry(ctx context.Context, id int64) error
        +DeleteSource(ctx context.Context, id int64) error
    }

    class XMLParser {
        +Parse(file io.Reader) ([]DictTerm, error)
    }

    class SQLiteDictionaryStore {
        -db *sql.DB
        +SaveTerms(ctx context.Context, terms []DictTerm) error
    }

    class DictionaryImporterImpl {
        -XMLParser parser
        -DictionaryStore store
        +ImportXML(ctx context.Context, file io.Reader) (int, error)
    }

    DictionaryService --> DictionaryImporter : uses
    DictionaryService --> DictionaryStore : uses
    DictionaryImporter <|.. DictionaryImporterImpl : implements
    DictionaryImporterImpl --> XMLParser : uses
    DictionaryImporterImpl --> DictionaryStore : uses
    DictionaryStore <|.. SQLiteDictionaryStore : implements
    DictionaryImporterImpl ..> DictTerm : creates
    DictionaryStore ..> DictTerm : stores
```

## アーキテクチャの補足：基本インフラの注入による純粋な Vertical Slicing
本コンテキスト（Dictionary Slice）は、**「XMLパース」から「DBテーブルスキーマ(DTO)定義」「SQL永続化」までの全責務をこのスライス単体で負う**。
AIDDにおいてAIが変更範囲を迷わず限定・自己完結させて決定的にコードを生成できるよう、あえて全体での「DRY」は捨て、他のコンテキスト（例：翻訳時の辞書読み込み等）とはStoreやモデルを共有しない。
外部（プロセスマネージャー等）からは、DBのプーリングや接続管理のためだけのインフラモジュール（例：`*sql.DB` コネクションプール）のみをDIで注入する形とする。

## 推奨ライブラリ (Go Backend)
*   **XML 解析**: `encoding/xml` (標準ライブラリ)
*   **DB アクセス**: `github.com/mattn/go-sqlite3` (デファクトスタンダード、ただしCGO有効化が必要) または `modernc.org/sqlite` (CGO不要な代替品)
*   **依存性注入**: `github.com/google/wire` (プロジェクト標準)
*   **通信方式**: Wails バインディング (フロントエンドとの直接通信)
