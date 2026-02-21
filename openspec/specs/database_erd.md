# データベースER図

Interface-First AIDD v2アーキテクチャおよびVertical Slice Architecture (VSA)に従った、各Sliceのデータベース設計を以下に示します。各Sliceは独自の責務範囲に応じて、独立したコンテキスト（ファイル・テーブル群）として管理されます。

## Config Store Slice (設定・レイアウト保存)

共通の設定やUI状態を永続化するインフラストラクチャ層のコンテキストです。
**データベース名:** `config.db` (システム設定・全Mod共通)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#7a81ff',
  'lineColor': '#888888',
  'secondaryColor': '#2d3748',
  'tertiaryColor': '#4a5568',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    config ||--o| schema_version : "schema"
    ui_state ||--o| schema_version : "schema"
    secrets ||--o| schema_version : "schema"
    
    config {
        TEXT namespace PK "名前空間 (e.g. llm, dictionary_builder)"
        TEXT key PK "設定キー (e.g. selected_provider)"
        TEXT value "プレーン文字列（JSON不可）"
        DATETIME updated_at "最終更新日時"
    }

    ui_state {
        TEXT namespace PK "名前空間 (e.g. ui.layout)"
        TEXT key PK "設定キー"
        TEXT value "JSON形式の値（構造化データ可）"
        DATETIME updated_at "最終更新日時"
    }

    secrets {
        TEXT namespace PK "名前空間 (e.g. llm.gemini)"
        TEXT key PK "シークレットキー (e.g. api_key)"
        TEXT value "機密値（暗号化想定）"
        DATETIME updated_at "最終更新日時"
    }
    
    schema_version {
        INTEGER version "現在のスキーマバージョン"
        DATETIME applied_at "適用日時"
    }
```

## Dictionary Builder Slice (辞書構築)

公式DLCや基本辞書など、xTranslatorフォーマットから構築される汎用辞書データのコンテキストです。
**データベース名:** `dictionary.db` (システム辞書・全Mod共通)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#4caf50',
  'lineColor': '#888888',
  'secondaryColor': '#1b5e20',
  'tertiaryColor': '#2e7d32',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    %% Dictionary Builder Slice
    dlc_dictionary_entries {
        INTEGER id PK "自動採番ID"
        TEXT edid "Editor ID"
        TEXT record_type "レコードタイプ (e.g. BOOK:FULL)"
        TEXT source_text "原文(英語)"
        TEXT dest_text "翻訳文(日本語)"
    }
```

## NPC Persona Gen Slice (ペルソナ生成)

NPCの会話履歴から生成された性格や口調のペルソナ情報を管理するコンテキストです。
**データベース名:** `{PluginName}_terms.db` (生成元Mod専用データベース・テーブル同居)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#ff9800',
  'lineColor': '#888888',
  'secondaryColor': '#e65100',
  'tertiaryColor': '#ef6c00',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    %% Persona Gen Slice
    npc_personas {
        INTEGER id PK "自動採番ID"
        TEXT speaker_id "NPC識別子 (SpeakerID) UNIQUE"
        TEXT editor_id "NPC Editor ID"
        TEXT npc_name "NPC名"
        TEXT race "種族"
        TEXT sex "性別"
        TEXT voice_type "声の種類"
        TEXT persona_text "生成されたペルソナテキスト"
        INTEGER dialogue_count "ペルソナ生成に使用した会話件数"
        INTEGER estimated_tokens "推定トークン利用量"
        TEXT source_plugin "ソースプラグイン名"
        DATETIME created_at "作成日時"
        DATETIME updated_at "更新日時"
    }
```

## Term Translator Slice (Mod用語翻訳)

対象Mod固有の固有名詞翻訳結果と、その部分一致検索用のFTS（全文検索）テーブルを管理するコンテキストです。
**データベース名:** `{PluginName}_terms.db` (翻訳対象Mod専用データベース)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#e91e63',
  'lineColor': '#888888',
  'secondaryColor': '#880e4f',
  'tertiaryColor': '#ad1457',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    mod_terms ||--o{ mod_terms_fts : "FTS5同期 (トリガー)"
    mod_terms ||--o{ npc_terms_fts : "FTS5同期 (NPC用)"

    mod_terms {
        INTEGER id PK "自動採番ID"
        TEXT original_en "原文 (英語, 小文字正規化) UNIQUE(with record_type)"
        TEXT translated_ja "翻訳結果 (日本語)"
        TEXT record_type "レコードタイプ (e.g. NPC_ FULL) UNIQUE(with original_en)"
        TEXT editor_id "Editor ID"
        TEXT source_plugin "ソースプラグイン名"
        DATETIME created_at "作成日時"
    }

    mod_terms_fts {
        TEXT original_en "原文 (全文検索用 FTS5)"
        TEXT translated_ja "翻訳結果 (全文検索用 FTS5)"
    }

    npc_terms_fts {
        TEXT original_en "原文 (NPC専用部分一致検索用 FTS5)"
        TEXT translated_ja "翻訳結果 (NPC専用部分一致検索用 FTS5)"
    }
```

## 補足事項
- **Vertical Slice Architecture の原則**: VSAの原則（Refactoring Strategy Section 5）に基づき、上記テーブルはDRY原則を避け「あえて分断」されています。各Slice（`ConfigStoreSlice`, `DictionaryBuilderSlice`, `PersonaGenSlice`, `TermTranslatorSlice`）は自身が必要とするテーブルのみに依存し、他Sliceのテーブルに直接クエリを発行することはありません。
