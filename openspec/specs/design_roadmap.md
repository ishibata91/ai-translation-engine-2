# 設計ロードマップ

本ドキュメントは、Skyrim翻訳エンジン v2 (Go/React版) の設計フェーズにおける成果物とその作成順序を定義します。

## 1. コア・アーキテクチャ (基盤)
**目的**: 全体的な構造とデータの流れを定義する。
- **ドキュメント**: `architecture_overview.md`, `infrastructure/llm_client_interface.md`, `infrastructure/config_store/spec.md`
- **内容**:
    - **2-Pass System**: Propose/Translateの2段階翻訳戦略。
    - **Tech Stack**: Go (Backend) + React (Frontend) の責務分担。
    - **Data Flow**: xEdit (JSON) -> Go App -> LLM -> Output (JSON) のデータフロー。
    - **LLM Infrastructure**: 全スライス共通のLLMクライアントインターフェース定義。
    - **Config Store**: APIキー・LLM選択状態・UIレイアウト等をSQLiteで永続化する共通インフラ。

## 2. データスキーマ & ローダー (Phase 1)
**目的**: 堅牢な型定義を行い、データ基盤を安定させる。
- **ドキュメント**: `data_loader_design.md`
- **内容**:
    - **Class Diagram**: `extractData.pas` のJSON出力に対応するGo Struct定義。
    - **Sequence Diagram**: ファイル読み込み、バリデーション、メモリストアへの格納フロー。
    - **Error Handling**: 不正なデータ構造への対処方針。

## 3. 用語翻訳ロジック (Phase 2)
**目的**: 固有名詞の抽出と辞書管理の仕組みを設計する。
- **ドキュメント**: `term_translator/spec.md`, `dictionary_builder/spec.md`
- **内容**:
    - **Dictionary Builder**: xTranslator形式XMLからの用語抽出と、SQLiteベースの辞書DB構築を行う自律的Sliceの設計。
    - **Extraction Rules**: 正規表現やロジックによる固有名詞（NPC名、地名、アイテム名）の抽出ルール。
    - **Cache Schema**: 辞書ファイルの構造 (JSON/KV Store / SQLite)。
    - **Conflict Resolution**: 同義語や訳揺れの解決策。

## 4. 文脈エンジンロジック (Phase 3)
**目的**: AIに渡すコンテキスト構築ロジックを設計する。
- **ドキュメント**: `context_engine_design.md`, `npc_persona_generator/spec.md`
- **内容**:
    - **Context Builder**: 会話ツリー (`Previous Lines`) のトラバース方法。
    - **Speaker Profiling**: 種族・クラス・性格データからの口調 (`Tone`) 推定ロジック。
    - **NPC Persona Generator**: NPCごとに会話データ（最大100件）を収集し、LLMにリクエストしてペルソナ（性格・口調・背景）を自動生成するSliceの設計。NPCごとの想定トークン利用量を事前計算し、コンテキスト長の評価を行うことで、LLMのコンテキストウィンドウ超過を防止する。
    - **Prompt Engineering**: LLMへのシステムプロンプトのテンプレート設計。

## 5. Web API & UI (Phase 4)
**目的**: ユーザーインターフェースとAPIを設計する。
- **ドキュメント**: `api_ui_design.md`
- **内容**:
    - **API Definition**: ファイルロード、翻訳実行、設定変更などのREST/WebSocket API定義。
    - **UI Wireframes**: ダッシュボード、進捗表示、設定画面のモックアップ。全てのLLMユースケースで共通のLLM選択モーダルUI（ローカル/Gemini/xAI/バッチAPI）を含む。
    - **State Management**: React側での状態管理 (React Query / Context) 設計。

---

## 実行順序
以下の順序で設計ドキュメントを作成し、各段階でレビューを行います。

1.  **Data Schema** (完了)
2.  **Term Logic** (現在ここに着手)
3.  **Context Logic**
4.  **UI/API**
