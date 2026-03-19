# データベース ER 図

Interface-First AIDD v2アーキテクチャおよびVertical Slice Architecture (VSA)に従った、各Sliceのデータベース設計を以下に示します。各Sliceは独自の責務範囲に応じて、独立したコンテキスト（ファイル・テーブル群）として管理されます。

## DB配置方針

- すべてのSQLite DBは **プロセスのカレントディレクトリ（ワークスペースルート）配下の `db/`** に配置する。
- 例: ワークスペースが `F:\ai translation engine 2` の場合、DBは `F:\ai translation engine 2\db\*.db`。
- VSA原則に従い、スライスごとにDBファイルを分離する（例: `config.db`, `dictionary.db`, `task.db`）。
- 複数 slice から参照する shared data / handoff data は `artifact` 専用ストアへ配置し、slice ローカル DB から直接参照させない。

## artifact (slice間共有成果物)

slice 間受け渡しで使う共有データ、中間成果物、resume 用状態を保存するコンテキストです。
**データベース名:** `artifact.db` (shared artifact 専用)

artifact は汎用 `artifact_records` へ JSON blob を保存するのではなく、slice ごとに必要な構造化テーブル群を配置する。翻訳フロー入力では `task_id` を親キーとして parser の DTO 構造に対応するテーブル群を持つ。
また、translation flow から再利用される shared dictionary と master persona の正本は artifact ストアへ配置する。

```mermaid
erDiagram
    translation_input_tasks ||--o{ translation_input_files : "1 task = N files"
    translation_input_files ||--o{ translation_input_dialogue_groups : "1 file = N dialogue groups"
    translation_input_dialogue_groups ||--o{ translation_input_dialogue_responses : "1 group = N responses"
    translation_input_files ||--o{ translation_input_quests : "1 file = N quests"
    translation_input_quests ||--o{ translation_input_quest_stages : "1 quest = N stages"
    translation_input_quests ||--o{ translation_input_quest_objectives : "1 quest = N objectives"
    translation_input_files ||--o{ translation_input_items : "1 file = N items"
    translation_input_files ||--o{ translation_input_magic : "1 file = N magic"
    translation_input_files ||--o{ translation_input_locations : "1 file = N locations"
    translation_input_files ||--o{ translation_input_cells : "1 file = N cells"
    translation_input_files ||--o{ translation_input_system_records : "1 file = N system records"
    translation_input_files ||--o{ translation_input_messages : "1 file = N messages"
    translation_input_files ||--o{ translation_input_load_screens : "1 file = N load screens"
    translation_input_files ||--o{ translation_input_npcs : "1 file = N npcs"
    artifact_dictionary_sources ||--o{ artifact_dictionary_entries : "1 source = N entries"
    artifact_master_persona_final ||--o{ artifact_master_persona_temp : "lookup key を共有（物理FKなし）"

    translation_input_tasks {
        TEXT task_id PK "task.db の tasks.id と同じ値"
        TEXT status "pending / loaded / failed"
        DATETIME created_at "作成日時"
        DATETIME updated_at "更新日時"
    }

    translation_input_files {
        INTEGER id PK "ファイル行ID"
        TEXT task_id "translation_input_tasks.task_id に論理紐付け"
        TEXT source_file_path "入力ファイル絶対パス"
        TEXT source_file_name "表示用ファイル名"
        TEXT source_file_hash "重複検出用ハッシュ"
        TEXT parse_status "pending / loaded / failed"
        INTEGER preview_row_count "preview 対象合計件数 (保存時に確定)"
        DATETIME parsed_at "パース完了日時"
        DATETIME created_at "作成日時"
        DATETIME updated_at "更新日時"
    }

    translation_input_dialogue_groups {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id "BaseExtractedRecord.ID"
        TEXT editor_id "BaseExtractedRecord.EditorID"
        TEXT record_type "BaseExtractedRecord.Type"
        TEXT source_json_path "BaseExtractedRecord.SourceJSON"
        TEXT player_text
        TEXT quest_id
        INTEGER is_services_branch
        TEXT services_type
        TEXT nam1
        TEXT source
    }

    translation_input_dialogue_responses {
        INTEGER id PK
        INTEGER dialogue_group_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT text
        TEXT prompt
        TEXT topic_text
        TEXT menu_display_text
        TEXT speaker_id
        TEXT voice_type
        INTEGER response_order
        TEXT previous_id
        TEXT source
        INTEGER response_index
    }

    translation_input_quests {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT source
    }

    translation_input_quest_stages {
        INTEGER id PK
        INTEGER quest_id FK
        INTEGER stage_index
        INTEGER log_index
        TEXT stage_type
        TEXT text
        TEXT parent_id
        TEXT parent_editor_id
        TEXT source
    }

    translation_input_quest_objectives {
        INTEGER id PK
        INTEGER quest_id FK
        TEXT objective_index
        TEXT objective_type
        TEXT text
        TEXT parent_id
        TEXT parent_editor_id
        TEXT source
    }

    translation_input_items {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT description
        TEXT text
        TEXT type_hint
        TEXT source
    }

    translation_input_magic {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT description
        TEXT source
    }

    translation_input_locations {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT parent_id
        TEXT source
    }

    translation_input_cells {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT parent_id
        TEXT source
    }

    translation_input_system_records {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT description
        TEXT source
    }

    translation_input_messages {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT text
        TEXT title
        TEXT quest_id
        TEXT source
    }

    translation_input_load_screens {
        INTEGER id PK
        INTEGER file_id FK
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT text
        TEXT source
    }

    translation_input_npcs {
        INTEGER id PK
        INTEGER file_id FK
        TEXT npc_key "ParserOutput.NPCs の map key"
        TEXT source_record_id
        TEXT editor_id
        TEXT record_type
        TEXT source_json_path
        TEXT name
        TEXT race
        TEXT voice
        TEXT sex
        TEXT class_name
        TEXT source
    }

    artifact_dictionary_sources {
        INTEGER id PK "自動採番ID"
        TEXT file_name
        TEXT format
        TEXT file_path
        INTEGER file_size
        INTEGER entry_count
        TEXT status "PENDING / IMPORTING / COMPLETED / ERROR"
        TEXT error_message
        DATETIME imported_at
        DATETIME created_at
    }

    artifact_dictionary_entries {
        INTEGER id PK "自動採番ID"
        INTEGER source_id FK "artifact_dictionary_sources.id"
        TEXT edid
        TEXT record_type
        TEXT source_text
        TEXT dest_text
    }

    artifact_master_persona_final {
        INTEGER persona_id PK "自動採番ID"
        TEXT form_id "UI 表示用 FormID"
        TEXT source_plugin "lookup key (unique with speaker_id)"
        TEXT speaker_id "lookup key (unique with source_plugin)"
        TEXT npc_name
        TEXT editor_id
        TEXT race
        TEXT sex
        TEXT voice_type
        DATETIME updated_at
        TEXT persona_text "最終成果物テキスト"
        TEXT generation_request "UI表示用リクエスト本文"
        TEXT dialogues_json "UI表示用 dialogue 一覧(JSON)"
    }

    artifact_master_persona_temp {
        INTEGER id PK "自動採番ID"
        TEXT task_id "task スコープ cleanup キー"
        TEXT source_plugin "lookup key"
        TEXT speaker_id "lookup key"
        TEXT editor_id
        TEXT npc_name
        TEXT race
        TEXT sex
        TEXT voice_type
        TEXT generation_request "中間生成物"
        TEXT dialogues_json "中間生成物"
        DATETIME updated_at
    }
```

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
    dlc_sources ||--o{ dlc_dictionary_entries : "1ソース = N エントリ"

    dlc_sources {
        INTEGER id PK "自動採番ID"
        TEXT file_name "ファイル名 (e.g. Skyrim.esm) UNIQUE"
        TEXT format "フォーマット (e.g. SSTXML)"
        TEXT file_path "インポート元ファイルの絶対パス"
        INTEGER file_size_bytes "ファイルサイズ (bytes)"
        INTEGER entry_count "インポートされたエントリ数 (完了後に更新)"
        TEXT status "インポート状態 (PENDING / IMPORTING / COMPLETED / ERROR)"
        TEXT error_message "エラーメッセージ (ERROR時のみ, nullable)"
        DATETIME imported_at "インポート完了日時 (nullable)"
        DATETIME created_at "レコード作成日時"
    }

    dlc_dictionary_entries {
        INTEGER id PK "自動採番ID"
        INTEGER source_id FK "dlc_sources.id (インポート元ソース)"
        TEXT edid "Editor ID"
        TEXT record_type "レコードタイプ (e.g. BOOK:FULL)"
        TEXT source_text "原文(英語)"
        TEXT dest_text "翻訳文(日本語)"
    }
```

### テーブル設計の補足

| テーブル                 | 画面対応                                         | 変更頻度                          |
| ------------------------ | ------------------------------------------------ | --------------------------------- |
| `dlc_sources`            | 辞書構築画面①「登録済みソース一覧」              | ファイルインポート時のみ (低頻度) |
| `dlc_dictionary_entries` | 辞書構築画面②「エントリ一覧」・③「エントリ編集」 | エントリ手動修正時 (中頻度)       |

> [!NOTE]
> `dlc_sources.entry_count` は `dlc_dictionary_entries` の行数と冗長になるが、一覧表示でのカウントクエリを省略するためのキャッシュカラムとして許容する。

## persona (ペルソナ生成)

NPCの会話履歴から生成された性格や口調のペルソナ情報を管理するコンテキストです。
**データベース名:** `persona.db` (ペルソナスライス専用・全Mod共通)

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
    npc_personas ||--o{ npc_dialogues : "1 NPC = N dialogues"

    npc_personas {
        INTEGER id PK "自動採番ID"
        TEXT speaker_id "NPC識別子 (SpeakerID)"
        TEXT editor_id "NPC Editor ID"
        TEXT npc_name "NPC名"
        TEXT race "種族"
        TEXT sex "性別"
        TEXT voice_type "声の種類"
        TEXT persona_text "生成されたペルソナテキスト"
        TEXT generation_request "LLMへ送信した生成リクエスト本文"
        TEXT status "状態 (draft / generated)"
        TEXT source_plugin "ソースプラグイン名 (UNIQUE with speaker_id)"
        DATETIME updated_at "更新日時"
    }

    npc_dialogues {
        INTEGER id PK "自動採番ID"
        INTEGER persona_id FK "npc_personas.id"
        TEXT source_plugin "ソースプラグイン名"
        TEXT speaker_id "NPC識別子 (検索補助)"
        TEXT editor_id "会話レコードのEditor ID"
        TEXT record_type "会話レコード種別"
        TEXT source_text "原文"
        TEXT quest_id "関連Quest ID (nullable)"
        INTEGER is_services_branch "services分岐フラグ (0/1)"
        INTEGER dialogue_order "元データ上の並び順"
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

インフラ層の汎用ジョブキュー。ドメイン知識を一切持たず、`process_id` / `task_id` / `task_type` と `request` のペアを永続化し、request単位の再開状態を保持します。`Completed` になった MasterPersona task の job は `task_id` 単位で削除されます。
**データベース名:** `llm_queue.db` (インフラ専用・全Mod共通)

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
        TEXT task_id "タスクID (UUID, 再開単位)"
        TEXT task_type "タスク種別 (e.g. persona_extraction)"
        TEXT request_json "LLMリクエスト (JSON)"
        TEXT status "PENDING / IN_PROGRESS / COMPLETED / FAILED / CANCELLED"
        TEXT request_state "pending / running / completed / failed / canceled"
        INTEGER resume_cursor "再開位置カーソル (未進行は0)"
        TEXT provider "実行プロバイダ (nullable)"
        TEXT model "実行モデル (nullable)"
        TEXT request_fingerprint "リクエスト一意性ハッシュ (nullable)"
        TEXT structured_output_schema_version "構造化出力スキーマ版 (nullable)"
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

## translator (本文翻訳結果)

Pass 2: 本文翻訳のスライスが管理する翻訳結果のコンテキストです。
**データベース名:** `{PluginName}_translations.db` (ソースプラグイン別)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#2196f3',
  'lineColor': '#888888',
  'secondaryColor': '#0d47a1',
  'tertiaryColor': '#1565c0',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    main_translations {
        INTEGER id PK "自動採番ID"
        TEXT form_id "FormID"
        TEXT record_type "レコードタイプ (e.g. INFO NAM1)"
        TEXT source_text "原文 (英語)"
        TEXT translated_text "翻訳結果 (日本語, nullable)"
        INTEGER stage_index "Stage/Objective Index (nullable)"
        TEXT status "処理状態 (success, failed, skipped, cached)"
        TEXT error_message "エラーメッセージ (nullable)"
        TEXT source_plugin "ソースプラグイン名"
        TEXT editor_id "Editor ID (nullable)"
        TEXT parent_form_id "親のFormID (nullable)"
        TEXT parent_editor_id "親のEditorID (nullable)"
        DATETIME created_at "作成日時"
        DATETIME updated_at "更新日時"
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

## frontend_tasks (UIタスク・進捗管理)

フロントエンドと連携する非同期タスク（辞書構築、ペルソナ抽出、翻訳プロジェクト等）のメタデータと進捗状態を永続化し、クラッシュリカバーやフェーズに応じた画面遷移を可能にするコンテキストです。
**データベース名:** `task.db` (タスクスライス専用データベース)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': {
  'primaryColor': '#2b2b2b',
  'primaryTextColor': '#e0e0e0',
  'primaryBorderColor': '#ffeb3b',
  'lineColor': '#888888',
  'secondaryColor': '#fbc02d',
  'tertiaryColor': '#f57f17',
  'mainBkg': '#1e1e1e',
  'nodeBorder': '#5a5a5a'
}}}%%
erDiagram
    tasks ||--o| schema_version : "schema"

    tasks {
        TEXT id PK "タスクID (UUID)"
        TEXT name "タスク名/タイトル"
        TEXT task_type "タスクの種別 (e.g. dictionary_import, translation_project)"
        TEXT status "全体ステータス (PENDING/RUNNING/PAUSED/COMPLETED/FAILED/CANCELLED)"
        TEXT phase "現在のフェーズ/ステップ (UI状態復元用)"
        INTEGER progress "全体進捗 (0-100)"
        TEXT metadata_json "設定値や再開用パラメータ (JSON)"
        TEXT error_message "エラーメッセージ (nullable)"
        DATETIME created_at "登録日時"
        DATETIME updated_at "最終更新日時"
    }

    schema_version {
        INTEGER version "現在のスキーマバージョン"
        DATETIME applied_at "適用日時"
    }
```

## 補足事項
- **Vertical Slice Architecture の原則**: VSAの原則（Architecture Section 5）に基づき、上記テーブルはDRY原則を避け「あえて分断」されています。各Slice（`config`, `dictionary`, `persona`, `terminology`）は自身が必要とするテーブルのみに依存し、他Sliceのテーブルに直接クエリを発行することはありません。
