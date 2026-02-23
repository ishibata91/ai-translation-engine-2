## MODIFIED Requirements

### Dictionary Slice (旧 Dictionary Builder)

公式DLCや基本辞書など、xTranslatorフォーマットから構築される汎用辞書データのコンテキストです。
**データベース名:** `dictionary.db`

```mermaid
erDiagram
    dlc_dictionary_entries {
        INTEGER id PK "自動採番ID"
        TEXT edid "Editor ID"
        TEXT record_type "レコードタイプ (e.g. BOOK:FULL)"
        TEXT source_text "原文(英語)"
        TEXT dest_text "翻訳文(日本語)"
    }
```

### Terminology Slice (旧 Term Translator)

対象Mod固有の固有名詞翻訳結果と、その部分一致検索用のFTS（全文検索）テーブルを管理するコンテキストです。
**データベース名:** `{PluginName}_terms.db`

```mermaid
erDiagram
    mod_terms ||--o{ mod_terms_fts : "FTS5同期 (トリガー)"
    mod_terms ||--o{ npc_terms_fts : "FTS5同期 (NPC用)"

    mod_terms {
        INTEGER id PK "自動採番ID"
        TEXT original_en "原文 (英語) UNIQUE(with record_type)"
        TEXT translated_ja "翻訳結果 (日本語)"
        TEXT record_type "レコードタイプ (e.g. NPC_ FULL)"
        TEXT editor_id "Editor ID"
        TEXT source_plugin "ソースプラグイン名"
        DATETIME created_at "作成日時"
    }
```

### Pipeline Slice (旧 Process Manager)

各スライスの実行状態やJobQueueとの紐付けを管理し、プロセスのレジューム（再開）を可能にするコンテキストです。
**データベース名:** `pipeline.db`

#### Scenario: パイプラインのレジューム
- **WHEN** 処理が途中で中断され、再起動される
- **THEN** `pipeline` スライスがDBから前回の到達フェーズを読み取り、実行を再開する。

### 補足事項
- **Vertical Slice Architecture の原則**: VSAの原則（`specs/architecture.md` 参照）に基づき、各テーブルは独立して管理されます。各Sliceは自身が必要とするテーブルのみに依存し、他Sliceのテーブルに直接クエリを発行することはありません。
