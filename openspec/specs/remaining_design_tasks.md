# 残り設計タスク一覧 (Remaining Design Tasks)

`openspec/specs/design_roadmap.md` および `openspec/specs/requirements.md` と現在の `specs` フォルダの状況を比較反映し、Skyrim翻訳エンジン v2 の今後の設計タスクを整理しました。

## フェーズの進捗概要
- **Phase 1 (Data Foundation):** 完了
- **Phase 2 (Term Logic):** ほぼ完了
- **Phase 3 (Context Logic):** 完了
- **Phase 4 (UI/API):** 未着手

---

## 🚀 未着手の設計タスク

### Phase 3: 文脈エンジンロジック (Context Logic)

#### 1. Context Engine Slice (✅ 完了)
LLMに渡すコンテキスト構築ロジック (`ContextEngineSlice/spec.md`) そのものの詳細化。
- **会話の流れ予測**: 会話ツリー（`Previous Lines`）トラバース方式の設計。
- **話者プロファイリング**: 種族・クラス・性格データから適切な口調 (`Tone`) を推定し適用するロジック（`requirements.md` §3.2）。
- **プロンプトエンジニアリングの基本テンプレート設計**: レコード種別に基づくテンプレート呼び出しとパラメータ埋め込みの設計。

### Phase 4: Web API & UI (API & UI Design)

#### 2. API / UI 全体設計 (api_ui_design.md)
ReactフロントエンドとGoバックエンド間を繋ぐAPIとUIの設計仕様書作成。
- **API Spec (REST/WebSocket)**:
    - ファイルロードの進行状況・状態の取得API
    - 翻訳ジョブの制御（開始、一時停止、再開、キャンセル）
    - 設定のCRUD操作 (Config Store へのアクセス)
- **UI State Management**:
    - React側でのグローバルな状態管理の設計方針 (React Query / Context API等)
    - リアルタイム進捗とログプレビューのデータフロー
- **プロンプトテンプレートエディタ**:
    - `ConfigStore` と連動したレコード種別ごとのシステムプロンプト編集画面のデータモデリング。

#### 3. LLM選択・設定モーダル設計
すべてのLLMプロバイダを統合管理するUIモジュールの設計（`requirements.md` §3.4）。
- **対象プロバイダ**: Gemini, xAI (Batch API含む), OpenAI, Local/GGUF。
- **データフロー**: モーダルから `ConfigStoreSlice` のAPIまたはDB更新を通じて状態を永続化し、稼働中の `LLMManager` に反映する仕組みの設計。

#### 4. Export Slice / Resume機能の設計
最終出力のJSON生成及び、中断回復機能の設計（`requirements.md` §3.5）。
- **差分更新（Resume）機構**:
    - 既存の翻訳結果をシームレスに読み込み、成功済み・エラー・未着手を判定して実行計画にマージするロジック。
- **FormIDの正規化**: 出力時、ロードオーダーに依存しないFormID形式（例: `0x001234|ModName.esp`）に統一して出力する戦略。
