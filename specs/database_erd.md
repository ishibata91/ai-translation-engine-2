# データベースER図

Interface-First AIDD v2アーキテクチャおよびVertical Slice Architecture (VSA)に従った、各Sliceのデータベース設計を以下に示します。各Sliceは独自の責務範囲に応じて、独立したコンテキスト（ファイル・テーブル群）として管理されます。

## config (設定・レイアウト保存)

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
        TEXT namespace PK "名前空間 (e.g. llm, dictionary)"
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

## dictionary (辞書構築)

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
    %% Dictionary Slice
    dlc_dictionary_entries {
        INTEGER id PK "自動採番ID"
        TEXT edid "Editor ID"
        TEXT record_type "レコードタイプ (e.g. BOOK:FULL)"
        TEXT source_text "原文(英語)"
        TEXT dest_text "翻訳文(日本語)"
    }
```

## persona (ペルソナ生成)

NPCの会話履歴から生成された性格や口調のペルソナ情報を管理するコンテキストです。
**データベース名:** `{PluginName}_persona.db` (生成元Mod専用データベース・テーブル同居)

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
    %% Persona Slice
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

## terminology (Mod用語翻訳)

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

## summary (要約キャッシュ)

会話やクエストの背景情報をLLMで要約し、再利用するためのキャッシュを管理するコンテキストです。
**データベース名:** `{PluginName}_summary_cache.db` (ソースプラグイン別)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#9c27b0',
  'lineColor': '#888888',
  'secondaryColor': '#4a148c',
  'tertiaryColor': '#6a1b9a',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    summaries {
        INTEGER id PK "自動採番ID"
        TEXT cache_key "ハッシュキー ({record_id}|{sha256_hash}) UNIQUE"
        TEXT summary_text "生成された要約文 (英語)"
        DATETIME updated_at "最終更新日時"
    }
```

## queue (LLMジョブキュー)

インフラ層の汎用ジョブキュー。ドメイン知識を一切持たず、`ProcessID` と `Request` のペアを永続化します。完了済みジョブはスライスへの結果渡し後に即時物理削除（Hard Delete）されるため、テーブルは常に最小サイズを維持します。
**データベース名:** `llm_jobs.db` (インフラ専用・全Mod共通)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#00bcd4',
  'lineColor': '#888888',
  'secondaryColor': '#006064',
  'tertiaryColor': '#00838f',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    llm_jobs ||--o| schema_version : "schema"

    llm_jobs {
        TEXT id PK "ジョブID (UUID)"
        TEXT process_id "処理単位ID (UUID, INDEX) ProcessManagerが刈り取りに使用"
        TEXT request_json "LLMリクエスト (JSON)"
        TEXT status "PENDING / IN_PROGRESS / COMPLETED / FAILED"
        TEXT batch_job_id "Batch API ジョブID (batch戦略時のみ使用, nullable)"
        TEXT response_json "完了時のLLMレスポンス (JSON, nullable)"
        TEXT error_message "エラーメッセージ (FAILED時のみ, nullable)"
        DATETIME created_at "登録日時"
        DATETIME updated_at "最終更新日時"
    }

    schema_version {
        INTEGER version "現在のスキーマバージョン"
        DATETIME applied_at "適用日時"
    }
```

## pipeline (進行状態管理)

各スライスの実行状態やJobQueueとの紐付けを管理し、プロセスのレジューム（再開）を可能にするコンテキストです。
**データベース名:** `pipeline.db` (管理用データベース)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#f44336',
  'lineColor': '#888888',
  'secondaryColor': '#b71c1c',
  'tertiaryColor': '#c62828',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    process_states ||--o| schema_version : "schema"

    process_states {
        TEXT id PK "プロセスID (UUID)"
        TEXT target_slice "対象スライス名"
        TEXT input_file "入力元ファイルパス/識別子"
        TEXT batch_job_id "Batch API ジョブID (nullable)"
        TEXT current_phase "現在のフェーズ (PREPARE/WAITING/SAVE)"
        TEXT status "全体ステータス (PENDING/IN_PROGRESS/COMPLETED/FAILED)"
        DATETIME updated_at "最終更新日時"
    }

    schema_version {
        INTEGER version "現在のスキーマバージョン"
        DATETIME applied_at "適用日時"
    }
```

## 補足事項
- **Vertical Slice Architecture の原則**: VSAの原則（Architecture Section 5）に基づき、上記テーブルはDRY原則を避け「あえて分断」されています。各Slice（`config`, `dictionary`, `persona`, `terminology`）は自身が必要とするテーブルのみに依存し、他Sliceのテーブルに直接クエリを発行することはありません。
